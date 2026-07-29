class Portfolio < Formula
  desc "Local-first project inventory and knowledge platform"
  homepage "https://github.com/shafi-/portfolio"

  if Hardware::CPU.arm?
    url "https://github.com/shafi-/portfolio/releases/latest/download/portfolio-darwin-arm64"
    sha256 :no_check
  else
    url "https://github.com/shafi-/portfolio/releases/latest/download/portfolio-darwin-amd64"
    sha256 :no_check
  end

  def install
    if Hardware::CPU.arm?
      bin.install "portfolio-darwin-arm64" => "portfolio"
    else
      bin.install "portfolio-darwin-amd64" => "portfolio"
    end
  end

  test do
    system "#{bin}/portfolio", "--version"
  end
end
