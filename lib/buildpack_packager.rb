require "base_packager"

class BuildpackPackager < BasePackager
  def dependencies
    [
      "http://go.googlecode.com/files/go1.1.1.linux-amd64.tar.gz",
      "http://go.googlecode.com/files/go1.1.2.linux-amd64.tar.gz",
      "http://go.googlecode.com/files/go1.1.linux-amd64.tar.gz",
      "http://go.googlecode.com/files/go1.2.1.linux-amd64.tar.gz",
      "http://go.googlecode.com/files/go1.2.linux-amd64.tar.gz",
    ]
  end

  def excluded_files
    [
      /^test-godir\b/
    ]
  end
end

BuildpackPackager.new("go", ARGV.first.to_sym).package
