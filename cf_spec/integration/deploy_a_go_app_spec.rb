$: << 'cf_spec'
require 'spec_helper'

describe 'CF Go Buildpack' do
  subject(:app) { Machete.deploy_app(app_name) }
  let(:browser) { Machete::Browser.new(app) }

  after { Machete::CF::DeleteApp.new.execute(app) }

  context 'with cached buildpack dependencies', :cached do
    context 'app has dependencies' do
      let(:app_name) { 'go_app_with_dependencies/src/go_app_with_dependencies' }

      specify do
        expect(app).to be_running
        expect(app).to have_logged('Hello from foo!')

        browser.visit_path('/')
        expect(browser).to have_body('hello, world')

        expect(app.host).not_to have_internet_traffic
      end
    end

    context 'app has vendored dependencies' do
      let(:app_name) { 'go_with_vendor_experiment_flag/src/go_app' }

      specify do
        expect(app).to be_running
        expect(app).to have_logged('Init: a.A == 1')

        browser.visit_path('/')
        expect(browser).to have_body('Read: a.A == 1')

        expect(app.host).not_to have_internet_traffic
      end
    end

    context 'app with vendored dependencies has Godeps.json with no Packages array' do
      let(:app_name) { 'go15vendorexperiment_no_packages_array/src/go_app' }

      specify do
        expect(app).to be_running
        expect(app).to have_logged('Init: a.A == 1')

        browser.visit_path('/')
        expect(browser).to have_body('Read: a.A == 1')

        expect(app.host).not_to have_internet_traffic
      end
    end

    context 'app has no dependencies' do
      let(:app_name) { 'go_app/src/go_app' }

      specify do
        expect(app).to be_running

        browser.visit_path('/')
        expect(browser).to have_body('go, world')
        expect(app).to have_logged(/Installing go[\d\.]+\.\.\. done/)
        expect(app).to have_logged(/Downloaded \[file:\/\/.*\]/)

        expect(app.host).not_to have_internet_traffic
      end
    end

    context 'expects a non-packaged version of go' do
      let(:app_name) { 'go99/src/go99' }
      let(:resource_url) { "https://storage.googleapis.com/golang/go99.99.99.linux-amd64.tar.gz" }

      it "displays useful understandable errors" do
        expect(app).not_to be_running

        expect(app).to have_logged(/failed/i)
        expect(app).to have_logged 'DEPENDENCY MISSING IN MANIFEST: go 99.99.99'
        expect(app).to_not have_logged 'Installing go99.99.99'
        expect(app).to_not have_logged('Uploading droplet')

        expect(app.host).not_to have_internet_traffic
      end
    end

    context 'heroku example' do
      let(:app_name) { 'go_heroku_example/src/go_heroku_example' }

      specify do
        expect(app).to be_running

        browser.visit_path('/')
        expect(browser).to have_body('hello, heroku')

        expect(app.host).not_to have_internet_traffic
      end
    end

    context 'a go app using ldflags' do
      context 'with version 1.5 or greater' do
        let(:app_name) { 'go1.5_app_using_ldflags/src/go_app' }

        specify do
          expect(app).to be_running
          browser.visit_path('/')
          expect(browser).to have_body('flag_linked')
          expect(app).to have_logged('main.linker_flag=flag_linked')
          expect(app.host).not_to have_internet_traffic
        end
      end
      context 'with version 1.4.2 or less' do
        let(:app_name) { 'go1.4.2_app_using_ldflags/src/go_app' }

        specify do
          expect(app).to be_running
          browser.visit_path('/')
          expect(browser).to have_body('flag_linked')
          expect(app.host).not_to have_internet_traffic
        end
      end
    end

  end

  context 'without cached buildpack dependencies', :uncached do
    context 'app has dependencies' do
      let(:app_name) { 'go_app_with_dependencies/src/go_app_with_dependencies' }

      specify do
        expect(app).to be_running
        expect(app).to have_logged('Hello from foo!')

        browser.visit_path('/')
        expect(browser).to have_body('hello, world')
      end
    end

    context 'app has vendored dependencies' do
      let(:app_name) { 'go_with_vendor_experiment_flag/src/go_app' }

      specify do
        expect(app).to be_running
        expect(app).to have_logged('Init: a.A == 1')

        browser.visit_path('/')
        expect(browser).to have_body('Read: a.A == 1')
      end
    end

    context 'app with vendored dependencies has Godeps.json with no Packages array' do
      let(:app_name) { 'go15vendorexperiment_no_packages_array/src/go_app' }

      specify do
        expect(app).to be_running
        expect(app).to have_logged('Init: a.A == 1')

        browser.visit_path('/')
        expect(browser).to have_body('Read: a.A == 1')
      end
    end

    context 'app has no dependencies' do
      let(:app_name) { 'go_app/src/go_app' }

      specify do
        expect(app).to be_running

        browser.visit_path('/')
        expect(browser).to have_body('go, world')
        expect(app).to have_logged(/Installing go[\d\.]+\.\.\. done/)
        expect(app).to have_logged(/Downloaded \[https:\/\/.*\]/)
      end
    end

    context 'expects a non-existent version of go' do
      let(:app_name) { 'go99/src/go99' }
      let(:resource_url) { "https://storage.googleapis.com/golang/go99.99.99.linux-amd64.tar.gz" }

      it "displays useful understandable errors" do
        expect(app).not_to be_running

        expect(app).to have_logged 'DEPENDENCY MISSING IN MANIFEST: go 99.99.99'
        expect(app).to_not have_logged 'Installing go99.99.99'
      end
    end

    context 'heroku example' do
      let(:app_name) { 'go_heroku_example/src/go_heroku_example' }

      specify do
        expect(app).to be_running

        browser.visit_path('/')
        expect(browser).to have_body('hello, heroku')
      end
    end

    context 'a go app using ldflags' do
      context 'with version 1.5 or greater' do
        let(:app_name) { 'go1.5_app_using_ldflags/src/go_app' }

        specify do
          expect(app).to be_running
          browser.visit_path('/')
          expect(browser).to have_body('flag_linked')
          expect(app).to have_logged('main.linker_flag=flag_linked')
        end
      end
      context 'with version 1.4.2 or less' do
        let(:app_name) { 'go1.4.2_app_using_ldflags/src/go_app' }

        specify do
          expect(app).to be_running
          browser.visit_path('/')
          expect(browser).to have_body('flag_linked')
        end
      end
    end
  end

  context 'a .godir file is detected' do
    let(:app_name) { 'go_deprecated_heroku_example/src/go_heroku_example' }

    it 'fails with a deprecation message' do
      expect(app).to_not be_running
      expect(app).to have_logged('Deprecated, .godir file found! Please update to supported Godeps dependency manager.')
      expect(app).to have_logged('See https://github.com/tools/godep for usage information.')
    end
  end

  context 'a go app with wildcard matcher' do
    let(:app_name) { 'go_app_with_wildcard_version/src/go_app' }

    specify do
      expect(app).to be_running
      browser.visit_path('/')
      expect(browser).to have_body('go, world')
      expect(app).to have_logged(/Installing go1.4.3\.\.\. done/)
    end
  end

  context 'a go app with an invalid wildcard matcher' do
    let(:app_name) { 'go_app_with_invalid_wildcard_version/src/go_app' }

    specify do
      expect(app).to_not be_running

      expect(app).to have_logged 'DEPENDENCY MISSING IN MANIFEST: go 1.3'
      expect(app).to_not have_logged 'Installing go1.3'
    end
  end

end
