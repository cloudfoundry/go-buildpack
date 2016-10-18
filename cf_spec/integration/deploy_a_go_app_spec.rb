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

        expect(app).not_to have_internet_traffic
      end

      context 'app uses go1.6 and godep with GO15VENDOREXPERIMENT=0' do
        subject(:app) do
          Machete.deploy_app('go1.6_app_with_dependencies/src/go_app_with_dependencies',
            env: {"GO15VENDOREXPERIMENT" => "0"})
        end

        specify do
          expect(app).to be_running
          expect(app).to have_logged('Hello from foo!')

          browser.visit_path('/')
          expect(browser).to have_body('hello, world')
        end
      end

      context 'app uses go1.6 and godep with Godeps/_workspace dir' do
        subject(:app) do
          Machete.deploy_app('go1.6_app_with_dependencies/src/go_app_with_dependencies')
        end

        specify do
          expect(app).to be_running
          expect(app).to have_logged('Hello from foo!')

          browser.visit_path('/')
          expect(browser).to have_body('hello, world')
        end
      end

      context 'app uses go1.6 with godep and no vendor dir or Godeps/_workspace dir' do
        subject(:app) do
          Machete.deploy_app('go1.6_app_with_no_vendor/src/go_app_with_dependencies')
        end

        specify do
          expect(app).to have_logged('vendor/ directory does not exist.')
        end
      end
    end

    context 'app has vendored dependencies' do
      let(:app_name) { 'go_with_vendor_experiment_flag/src/go_app' }

      specify do
        expect(app).to be_running
        expect(app).to have_logged('Init: a.A == 1')

        browser.visit_path('/')
        expect(browser).to have_body('Read: a.A == 1')

        expect(app).not_to have_internet_traffic
      end
    end

    context 'app has vendored dependencies and no Godeps folder' do
      let(:app_name) { 'go_with_native_vendoring/src/go_app' }
      subject(:app) do
        Machete.deploy_app(app_name, env: {'BP_DEBUG' => '1'})
      end

      it 'successfully stages' do
        expect(app).to be_running
        expect(app).to have_logged('Init: a.A == 1')

        browser.visit_path('/')
        expect(browser).to have_body('Read: a.A == 1')

        expect(app).not_to have_internet_traffic
      end

      it 'uses default_versions_for to pick the Go version' do
        expect(app).to have_logged('DEBUG: default_version_for go is')
      end
    end

    context 'app has vendored dependencies and custom package spec' do
      let(:app_name) { 'go_with_native_vendoring_custom_install_spec/src/go_app' }
      subject(:app) do
        Machete.deploy_app(app_name, env: {'BP_DEBUG' => '1'})
      end

      it 'successfully stages' do
        expect(app).to be_running
        expect(app).to have_logged('Init: a.A == 1')

        browser.visit_path('/')
        expect(browser).to have_body('Read: a.A == 1')

        expect(app).not_to have_internet_traffic
      end
    end

    context 'app has vendored dependencies and a vendor.json file' do
      let(:app_name) { 'go_with_native_vendoring_and_vendor_json/src/go_app' }

      it 'successfully stages' do
        expect(app).to be_running
        expect(app).to have_logged('Init: a.A == 1')

        browser.visit_path('/')
        expect(browser).to have_body('Read: a.A == 1')
      end
    end

    context 'app with only a single go file and GOPACKAGENAME specified' do
      let(:app_name) { 'go_single_file/src/go_app' }

      it 'successfully stages' do
        expect(app).to be_running

        browser.visit_path('/')
        expect(browser).to have_body('simple apps are good')
      end
    end

    context 'app with only a single go file, a vendor directory, and no GOPACKAGENAME specified' do
      let(:app_name) { 'go_with_native_vendoring_no_gopackagename/src/go_app' }

      it 'successfully stages' do
        expect(app).to_not be_running

        expect(app).to have_logged(/failed/i)
        expect(app).to have_logged 'To use go native vendoring set the $GOPACKAGENAME'
      end
    end

    context 'app has vendored dependencies with go1.6, but GO15VENDOREXPERIMENT=0' do
      let(:app_name) { 'go16_native_vendoring_bad_env/src/go_app' }

      it 'fails with helpful error' do
        expect(app).to_not be_running

        expect(app).to have_logged(/failed/i)
        expect(app).to have_logged 'for go 1.6 this environment variable must unset or set to 1'
      end
    end

    context 'app has vendored dependencies with go1.5, but GO15VENDOREXPERIMENT=0' do
      let(:app_name) { 'go15_native_vendoring_bad_env/src/go_app' }

      it 'fails with helpful error' do
        expect(app).to_not be_running

        expect(app).to have_logged(/failed/i)
        expect(app).to have_logged 'for go 1.5 this environment variable must be set to 1'
      end
    end

    context 'app with vendored dependencies has Godeps.json with no Packages array' do
      let(:app_name) { 'go15vendorexperiment_no_packages_array/src/go_app' }

      specify do
        expect(app).to be_running
        expect(app).to have_logged('Init: a.A == 1')

        browser.visit_path('/')
        expect(browser).to have_body('Read: a.A == 1')

        expect(app).not_to have_internet_traffic
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

        expect(app).not_to have_internet_traffic
      end
    end

    context 'app has no Procfile' do
      let(:app_name) { 'go_app_without_procfile/src/go_app_without_procfile' }

      specify do
        expect(app).to be_running

        browser.visit_path('/')
        expect(browser).to have_body('go, world')
        expect(app).to have_logged(/Installing go[\d\.]+\.\.\. done/)
        expect(app).to have_logged(/Downloaded \[file:\/\/.*\]/)

        expect(app).not_to have_internet_traffic
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
      end
    end

    context 'heroku example' do
      let(:app_name) { 'go_heroku_example/src/go_heroku_example' }

      specify do
        expect(app).to be_running

        browser.visit_path('/')
        expect(browser).to have_body('hello, heroku')

        expect(app).not_to have_internet_traffic
      end
    end

    context 'a go app using ldflags' do
      context 'with version 1.5~' do
        let(:app_name) { 'go1.5_app_using_ldflags/src/go_app' }

        specify do
          expect(app).to be_running
          browser.visit_path('/')
          expect(browser).to have_body('flag_linked')
          expect(app).to have_logged('main.linker_flag=flag_linked')
          expect(app).not_to have_internet_traffic
        end
      end

      context 'with version 1.6~' do
        let(:app_name) { 'go1.6_app_using_ldflags/src/go_app' }

        specify do
          expect(app).to be_running
          browser.visit_path('/')
          expect(browser).to have_body('flag_linked')
          expect(app).to have_logged('main.linker_flag=flag_linked')
          expect(app).not_to have_internet_traffic
        end
      end
    end

    context 'app uses glide and has vendored dependencies' do
      let(:app_name) { 'go_app_with_glide_and_vendoring/src/go_app_with_glide_and_vendoring' }

      specify do
        expect(app).to be_running
        expect(app).to have_logged('Hello from foo!')
        expect(app).to have_logged('Note: skipping (glide install) due to non-empty vendor directory.')

        browser.visit_path('/')
        expect(browser).to have_body('hello, world')

        expect(app).not_to have_internet_traffic
      end
    end

    context 'go 1.7 app with GO_SETUP_GOPATH_IN_IMAGE' do
      let(:app_name) { 'go_app_gopath_in_container/src/go_app' }
      subject(:app) do
        Machete.deploy_app(app_name, env: {'GO_SETUP_GOPATH_IN_IMAGE' => 'true'})
      end

      it 'displays the GOPATH' do
        expect(app).to be_running

        browser.visit_path('/')
        expect(browser).to have_body('GOPATH: /home/vcap/app')
      end
    end

    context 'go 1.7 app with GO_INSTALL_TOOLS_IN_IMAGE' do
      let(:app_name) { 'go_app_toolchain_in_container/src/go_app' }
      subject(:app) do
        Machete.deploy_app(app_name, env: {'GO_INSTALL_TOOLS_IN_IMAGE' => 'true'})
      end

      it 'displays the go version' do
        expect(app).to be_running

        browser.visit_path('/')
        expect(browser).to have_body('go version go1.7.1 linux/amd64')
      end
    end
  end

  context 'without cached buildpack dependencies', :uncached do
    context 'app uses glide' do
      let(:app_name) { 'go_app_with_glide/src/go_app_with_glide' }

      specify do
        expect(app).to be_running
        expect(app).to have_logged('Hello from foo!')

        browser.visit_path('/')
        expect(browser).to have_body('hello, world')
      end

      it "uses a proxy during staging if present" do
        expect(app).to use_proxy_during_staging
      end
    end

    context 'app has dependencies' do
      let(:app_name) { 'go_app_with_dependencies/src/go_app_with_dependencies' }

      specify do
        expect(app).to be_running
        expect(app).to have_logged('Hello from foo!')

        browser.visit_path('/')
        expect(browser).to have_body('hello, world')
      end

      it "uses a proxy during staging if present" do
        expect(app).to use_proxy_during_staging
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
      context 'with version 1.6~' do
        let(:app_name) { 'go1.6_app_using_ldflags/src/go_app' }

        specify do
          expect(app).to be_running
          browser.visit_path('/')
          expect(browser).to have_body('flag_linked')
          expect(app).to have_logged('main.linker_flag=flag_linked')
        end
      end
    end

    context 'app uses glide and has vendored dependencies' do
      let(:app_name) { 'go_app_with_glide_and_vendoring/src/go_app_with_glide_and_vendoring' }

      specify do
        expect(app).to be_running
        expect(app).to have_logged('Hello from foo!')

        browser.visit_path('/')
        expect(browser).to have_body('hello, world')
      end

      it "uses a proxy during staging if present" do
        expect(app).to use_proxy_during_staging
      end
    end
  end

  context 'a .godir file is detected' do
    let(:app_name) { 'go_deprecated_heroku_example/src/go_heroku_example' }

    it 'fails with a deprecation message' do
      expect(app).to_not be_running
      expect(app).to have_logged('Deprecated, .godir file found! Please update to supported Godep or Glide dependency managers.')
      expect(app).to have_logged('See https://github.com/tools/godep or https://github.com/Masterminds/glide for usage information.')
    end
  end

  context 'a go app with wildcard matcher' do
    let(:app_name) { 'go_app_with_wildcard_version/src/go_app' }

    specify do
      expect(app).to be_running
      browser.visit_path('/')
      expect(browser).to have_body('go, world')
      expect(app).to have_logged(/Installing go1\.6\.\d+\.\.\. done/)
    end
  end

  context 'a go 1.6 app' do
    let(:app_name) { 'go_16_app/src/go_app' }

    it 'should be compiled with buildmode=pie' do
      expect(app).to be_running
      browser.visit_path('/')
      expect(browser.body).to match(/foo: (.*)/)
      old_address = /foo: (.*)/.match(browser.body)[1]
      Machete::CF::RestartApp.new.execute(app)
      expect(app).to be_running
      browser.visit_path('/')
      expect(browser.body).to match(/foo: (.*)/)
      new_address = /foo: (.*)/.match(browser.body)[1]
      expect(new_address).not_to eq(old_address)
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
