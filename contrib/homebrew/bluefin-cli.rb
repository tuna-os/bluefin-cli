# bluefin-cli Homebrew formula — in-repo source of truth (tuna-os/bluefin-cli#15).
#
# Two delivery paths:
#
#   1. tuna-os/homebrew-tap (authoritative, automated): GoReleaser generates a
#      binary formula and publishes it to the tap's Formula/ directory on every
#      tagged release (see the `brews` block in .goreleaser.yaml and
#      docs/adr/0003-release-pipeline.md). No manual step beyond configuring
#      the HOMEBREW_TAP_SSH_KEY secret.
#
#   2. External taps that build from source, e.g.
#      ublue-os/homebrew-experimental-tap: copy this file into the tap's
#      Formula/ directory and keep it in sync on each release:
#
#        url     https://github.com/tuna-os/bluefin-cli/archive/refs/tags/v<tag>.tar.gz
#        sha256  curl -sL <url> | shasum -a 256
#
# The formula builds from source (like the other tuna-os formulas) so it works
# on macOS and Linux without prebuilt-binary checksums. The `-X` ldflag stamps
# the real version into `cmd.version` — without it `bluefin-cli --version`
# prints "dev" and the formula's test fails.
class BluefinCli < Formula
  desc "Bring the Bluefin terminal experience to any machine"
  homepage "https://github.com/tuna-os/bluefin-cli"
  url "https://github.com/tuna-os/bluefin-cli/archive/refs/tags/v0.9.6.tar.gz"
  sha256 "239fe9f071dfd55e35f65a1778f6ee0293542efacf86872eb7239e631e76ce92"
  license "Apache-2.0"
  head "https://github.com/tuna-os/bluefin-cli.git", branch: "main"

  depends_on "go" => :build

  def install
    ENV["CGO_ENABLED"] = "0"
    ldflags = "-s -w -X github.com/tuna-os/bluefin-cli/cmd.version=v#{version}"
    system "go", "build", *std_go_args(output: bin/"bluefin-cli", ldflags:)
    generate_completions_from_executable(bin/"bluefin-cli", "completion")
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/bluefin-cli --version")
  end
end
