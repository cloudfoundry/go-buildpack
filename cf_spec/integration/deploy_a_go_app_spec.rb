$: << 'cf_spec'
require "spec_helper"

describe 'deploying a go app', :go_buildpack do
  it "makes the homepage available" do
    Machete.deploy_app("go_app", :go) do |app|
      expect(app).to be_staged
      expect(app.homepage_html).to include "go, world"
    end
  end
end
