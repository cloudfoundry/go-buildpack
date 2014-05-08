require 'bundler/setup'
require 'machete'

RSpec.configure do |config|
  config.before(:suite) do
    Machete::BuildpackUploader.new(:go)
  end
end
