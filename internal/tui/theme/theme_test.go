package theme

// Tests for the semantic theme token package. The package had zero tests:
// Resolve maps the terminal's background + a Catppuccin flavor override to
// the semantic token set, and everything downstream (TUI styling) depends on
// those tokens being sane and the aliases staying in sync.

import (
	"image/color"
	"testing"

	"charm.land/lipgloss/v2"
)

// assertColor compares a color.Color against a hex string, e.g. "#1e1e2e".
func assertColor(t *testing.T, got color.Color, wantHex string) {
	t.Helper()
	want := lipgloss.Color(wantHex)
	gr, gg, gb, ga := got.RGBA()
	wr, wg, wb, wa := want.RGBA()
	if gr != wr || gg != wg || gb != wb || ga != wa {
		t.Errorf("color = %v, want %s", got, wantHex)
	}
}

func TestFlavorByName(t *testing.T) {
	tests := []struct {
		name   string
		flavor string
		wantOK bool
	}{
		{name: "latte", flavor: "latte", wantOK: true},
		{name: "frappe", flavor: "frappe", wantOK: true},
		{name: "frappe accented", flavor: "frappé", wantOK: true},
		{name: "macchiato", flavor: "macchiato", wantOK: true},
		{name: "mocha", flavor: "mocha", wantOK: true},
		{name: "auto is not a flavor", flavor: "auto", wantOK: false},
		{name: "empty", flavor: "", wantOK: false},
		{name: "unknown", flavor: "bogus", wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, ok := flavorByName(tt.flavor)
			if ok != tt.wantOK {
				t.Fatalf("flavorByName(%q) ok = %v, want %v", tt.flavor, ok, tt.wantOK)
			}
			if ok && f == nil {
				t.Fatalf("flavorByName(%q) returned nil flavor", tt.flavor)
			}
		})
	}
}

func TestFlavors(t *testing.T) {
	got := Flavors()
	want := []string{"auto", "latte", "frappe", "macchiato", "mocha"}
	if len(got) != len(want) {
		t.Fatalf("Flavors() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Flavors() = %v, want %v", got, want)
		}
	}
}

func TestResolve_AutoDarkUsesMocha(t *testing.T) {
	saved := Flavor
	defer func() { Flavor = saved }()
	Flavor = ""

	th := Resolve(true)
	if !th.IsDark {
		t.Error("IsDark = false, want true for dark terminal")
	}
	assertColor(t, th.Bg, "#1e1e2e")       // Mocha Base
	assertColor(t, th.Accent, "#89b4fa")   // Mocha Blue
	assertColor(t, th.Surface, "#313244")  // Mocha Surface0
	assertColor(t, th.TextBase, "#cdd6f4") // Mocha Text
}

func TestResolve_AutoLightUsesLatte(t *testing.T) {
	saved := Flavor
	defer func() { Flavor = saved }()
	Flavor = ""

	th := Resolve(false)
	if th.IsDark {
		t.Error("IsDark = true, want false for light terminal")
	}
	assertColor(t, th.Bg, "#eff1f5")       // Latte Base
	assertColor(t, th.Accent, "#1e66f5")   // Latte Blue
	assertColor(t, th.TextBase, "#4c4f69") // Latte Text
}

func TestResolve_ExplicitLatteOverridesDarkTerminal(t *testing.T) {
	saved := Flavor
	defer func() { Flavor = saved }()
	Flavor = "latte"

	th := Resolve(true) // dark terminal, but flavor pinned to latte
	if th.IsDark {
		t.Error("IsDark = true, want false when flavor is latte")
	}
	assertColor(t, th.Bg, "#eff1f5") // Latte Base, not Mocha
}

func TestResolve_ExplicitMochaOverridesLightTerminal(t *testing.T) {
	saved := Flavor
	defer func() { Flavor = saved }()
	Flavor = "mocha"

	th := Resolve(false) // light terminal, but flavor pinned to mocha
	if !th.IsDark {
		t.Error("IsDark = false, want true when flavor is mocha")
	}
	assertColor(t, th.Bg, "#1e1e2e")
}

func TestResolve_AllExplicitFlavors(t *testing.T) {
	saved := Flavor
	defer func() { Flavor = saved }()

	tests := []struct {
		flavor string
		bg     string
		dark   bool
	}{
		{flavor: "latte", bg: "#eff1f5", dark: false},
		{flavor: "frappe", bg: "#303446", dark: true},
		{flavor: "macchiato", bg: "#24273a", dark: true},
		{flavor: "mocha", bg: "#1e1e2e", dark: true},
	}
	for _, tt := range tests {
		t.Run(tt.flavor, func(t *testing.T) {
			Flavor = tt.flavor
			th := Resolve(false)
			assertColor(t, th.Bg, tt.bg)
			if th.IsDark != tt.dark {
				t.Errorf("IsDark = %v, want %v for flavor %q", th.IsDark, tt.dark, tt.flavor)
			}
		})
	}
}

func TestResolve_TokenAliasesStayInSync(t *testing.T) {
	saved := Flavor
	defer func() { Flavor = saved }()
	Flavor = ""

	th := Resolve(true)

	// Legacy aliases must mirror their canonical tokens.
	assertColor(t, th.PrimaryBorder, "#89b4fa") // == Accent
	assertColor(t, th.PrimaryText, "#cdd6f4")   // == TextBase
	assertColor(t, th.SecondaryText, "#bac2de") // == TextMuted (Mocha Subtext1)
	assertColor(t, th.FaintText, "#7f849c")     // == TextFaint
	assertColor(t, th.InvertedText, "#1e1e2e")  // == Bg
	assertColor(t, th.SuccessText, "#a6e3a1")   // == Success
	assertColor(t, th.WarningText, "#f9e2af")   // == Warning
	assertColor(t, th.ErrorText, "#f38ba8")     // == Error
	assertColor(t, th.InfoText, "#89dceb")      // == Info

	// Semantic tokens are distinct from one another.
	distinct := map[string]color.Color{
		"Bg": th.Bg, "Surface": th.Surface, "Overlay": th.Overlay,
		"Accent": th.Accent, "AccentAlt": th.AccentAlt,
		"Success": th.Success, "Warning": th.Warning, "Error": th.Error, "Info": th.Info,
	}
	seen := map[uint64]string{}
	for name, c := range distinct {
		r, g, b, a := c.RGBA()
		key := uint64(r)<<48 | uint64(g)<<32 | uint64(b)<<16 | uint64(a)
		if prev, dup := seen[key]; dup {
			t.Errorf("tokens %s and %s resolve to the same color", prev, name)
		}
		seen[key] = name
	}
}

func TestDefaultTheme(t *testing.T) {
	th := DefaultTheme
	if !th.IsDark {
		t.Error("DefaultTheme.IsDark = false, want true (dark fallback)")
	}
	if th.Accent == nil || th.Bg == nil {
		t.Error("DefaultTheme has nil tokens")
	}
}
