$: << 'cf_spec'
require 'spec_helper'
require 'open3'

describe 'CF Go Buildpack' do
  subject(:app) { Machete.deploy_app(app_name, deploy_options) }
  let(:browser) { Machete::Browser.new(app) }
  let(:deploy_options) { {} }

  after { Machete::CF::DeleteApp.new.execute(app) }

  context 'with cached buildpack dependencies', :cached do
    context 'app has dependencies' do
      let(:app_name) { 'with_dependencies/src/with_dependencies' }

      specify do
        expect(app).to be_running
        expect(app).to have_logged('Hello from foo!')

        browser.visit_path('/')
        expect(browser).to have_body('hello, world')

        expect(app).not_to have_internet_traffic
      end

      context 'app uses go1.6 and godep with GO15VENDOREXPERIMENT=0' do
        let(:app_name) { 'go16_dependencies/src/go16_dependencies' }
        let(:deploy_options) { { env: {"GO15VENDOREXPERIMENT" => "0"} } }

        specify do
          expect(app).to be_running
          expect(app).to have_logged('Hello from foo!')

          browser.visit_path('/')
          expect(browser).to have_body('hello, world')
        end
      end

      context 'app uses go1.6 and godep with Godeps/_workspace dir' do
        let(:app_name) { 'go16_dependencies/src/go16_dependencies' }

        specify do
          expect(app).to be_running
          expect(app).to have_logged('Hello from foo!')

          browser.visit_path('/')
          expect(browser).to have_body('hello, world')
        end
      end

      context 'app uses go1.6 with godep and no vendor dir or Godeps/_workspace dir' do
        let(:app_name) { 'go16_no_vendor/src/go16_no_vendor' }

        specify do
          expect(app).to have_logged('vendor/ directory does not exist.')
        end
      end
    end

    context 'app has vendored dependencies' do
      let(:app_name) { 'go17_vendor_experiment_flag/src/go_app' }

      specify do
        expect(app).not_to be_running
        expect(app).to have_logged('GO15VENDOREXPERIMENT is set, but is not supported by go1.7')
      end
    end

    context 'app has vendored dependencies and no Godeps folder' do
      let(:app_name) { 'native_vendoring/src/go_app' }

      it 'successfully stages' do
        expect(app).to be_running
        expect(app).to have_logged('Init: a.A == 1')

        browser.visit_path('/')
        expect(browser).to have_body('Read: a.A == 1')

        expect(app).not_to have_internet_traffic
      end
    end

    context 'app has vendored dependencies and custom package spec' do
      let(:app_name) { 'vendored_custom_install_spec/src/go_app' }
      let(:deploy_options) { { env: {'BP_DEBUG' => '1'} } }

      it 'successfully stages' do
        expect(app).to be_running
        expect(app).to have_logged('Init: a.A == 1')

        browser.visit_path('/')
        expect(browser).to have_body('Read: a.A == 1')

        expect(app).not_to have_internet_traffic
      end
    end

    context 'app has vendored dependencies and a vendor.json file' do
      let(:app_name) { 'with_vendor_json/src/go_app' }

      it 'successfully stages' do
        expect(app).to be_running
        expect(app).to have_logged('Init: a.A == 1')

        browser.visit_path('/')
        expect(browser).to have_body('Read: a.A == 1')
      end
    end

    context 'app with only a single go file and GOPACKAGENAME specified' do
      let(:app_name) { 'single_file/src/go_app' }

      it 'successfully stages' do
        expect(app).to be_running

        browser.visit_path('/')
        expect(browser).to have_body('simple apps are good')
      end
    end

    context 'app with only a single go file, a vendor directory, and no GOPACKAGENAME specified' do
      let(:app_name) { 'vendored_no_gopackagename/src/go_app' }

      it 'successfully stages' do
        expect(app).to_not be_running

        expect(app).to have_logged(/failed/i)
        expect(app).to have_logged 'To use go native vendoring set the $GOPACKAGENAME'
      end
    end

    context 'app has vendored dependencies with go1.6, but GO15VENDOREXPERIMENT=0' do
      let(:app_name) { 'go16_vendor_bad_env/src/go_app' }

      it 'fails with helpful error' do
        expect(app).to_not be_running

        expect(app).to have_logged(/failed/i)
        expect(app).to have_logged 'with go 1.6 this environment variable must unset or set to 1.'
      end
    end

    context 'app has no dependencies' do
      let(:app_name) { 'go_app/src/go_app' }

      specify do
        expect(app).to be_running

        browser.visit_path('/')
        expect(browser).to have_body('go, world')
        expect(app).to have_logged(/Installing go [\d\.]+/)
        expect(app).to have_logged(/Copy \[\/tmp\//)

        expect(app).not_to have_internet_traffic
      end
    end

    context 'app has before/after compile hooks' do
      let(:app_name) { 'go_app/src/go_app' }
      let(:deploy_options) { { env: {'BP_DEBUG' => '1'} } }

      it 'runs the hooks' do
        expect(app).to have_logged('HOOKS 1: BeforeCompile')
        expect(app).to have_logged('HOOKS 2: AfterCompile')

        expect(app).to be_running
        browser.visit_path('/')
        expect(browser).to have_body('go, world')
      end
    end

    context 'app has no Procfile' do
      let(:app_name) { 'no_procfile/src/no_procfile' }

      specify do
        expect(app).to be_running

        browser.visit_path('/')
        expect(browser).to have_body('go, world')
        expect(app).to have_logged(/Installing go [\d\.]+/)
        expect(app).to have_logged(/Copy \[\/tmp\//)

        expect(app).not_to have_internet_traffic
      end
    end

    context 'expects a non-packaged version of go' do
      let(:app_name) { 'go99/src/go99' }
      let(:resource_url) { "https://storage.googleapis.com/golang/go99.99.99.linux-amd64.tar.gz" }

      it "displays useful understandable errors" do
        expect(app).not_to be_running

        expect(app).to have_logged(/failed/i)
        expect(app).to have_logged 'Unable to determine Go version to install: no match found for 99.99.99'
        expect(app).to_not have_logged 'Installing go99.99.99'
        expect(app).to_not have_logged('Uploading droplet')
      end
    end

    context 'heroku example' do
      let(:app_name) { 'heroku_example/src/heroku_example' }

      specify do
        expect(app).to be_running

        browser.visit_path('/')
        expect(browser).to have_body('hello, heroku')

        expect(app).not_to have_internet_traffic
      end
    end

    context 'a go app using ldflags' do
      context 'with version 1.6~' do
        let(:app_name) { 'go16_ldflags/src/go_app' }

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
      let(:app_name) { 'glide_and_vendoring/src/glide_and_vendoring' }

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
      let(:app_name) { 'gopath_in_container/src/go_app' }
      let(:deploy_options) { { env: {'GO_SETUP_GOPATH_IN_IMAGE' => 'true'} } }

      it 'displays the GOPATH' do
        expect(app).to be_running

        browser.visit_path('/')
        expect(browser).to have_body('GOPATH: /home/vcap/app')
      end
    end

    context 'go 1.7 app with GO_INSTALL_TOOLS_IN_IMAGE' do
      let(:app_name) { 'toolchain_in_container/src/go_app' }
      let(:deploy_options) { { env: {'GO_INSTALL_TOOLS_IN_IMAGE' => 'true'} } }

      it 'displays the go version' do
        expect(app).to be_running

        browser.visit_path('/')
        expect(browser).to have_body('go version go1.7.6 linux/amd64')
      end

      context 'running a task' do
        before { skip_if_no_run_task_support_on_targeted_cf }

        it 'can find the specifed go in the container' do
          expect(app).to be_running

          Open3.capture2e('cf','run-task', 'go_app', 'echo "RUNNING A TASK: $(go version)"')[1].success? or raise 'Could not create run task'
          expect(app).to have_logged(/RUNNING A TASK: go version go1\.7\.6 linux\/amd64/)
        end
      end

      context 'and GO_SETUP_GOPATH_IN_IMAGE' do
        let(:deploy_options) { { env: {'GO_INSTALL_TOOLS_IN_IMAGE' => 'true', 'GO_SETUP_GOPATH_IN_IMAGE' => 'true'} } }

        it 'displays the go version' do
          expect(app).to be_running

          browser.visit_path('/')
          expect(browser).to have_body('go version go1.7.6 linux/amd64')

          browser.visit_path('/gopath')
          expect(browser).to have_body('GOPATH: /home/vcap/app')
        end
      end
    end
  end

  context 'without cached buildpack dependencies', :uncached do
    context 'app uses glide' do
      let(:app_name) { 'with_glide/src/with_glide' }

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
      let(:app_name) { 'with_dependencies/src/with_dependencies' }

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
      let(:app_name) { 'go17_vendor_experiment_flag/src/go_app' }

      specify do
        expect(app).not_to be_running
        expect(app).to have_logged('GO15VENDOREXPERIMENT is set, but is not supported by go1.7')
      end
    end

    context 'app has no dependencies' do
      let(:app_name) { 'go_app/src/go_app' }

      specify do
        expect(app).to be_running

        browser.visit_path('/')
        expect(browser).to have_body('go, world')
        expect(app).to have_logged(/Installing go [\d\.]+/)
        expect(app).to have_logged(/Download \[https:\/\/.*\]/)
      end
    end

    context 'expects a non-existent version of go' do
      let(:app_name) { 'go99/src/go99' }
      let(:resource_url) { "https://storage.googleapis.com/golang/go99.99.99.linux-amd64.tar.gz" }

      it "displays useful understandable errors" do
        expect(app).not_to be_running

        expect(app).to have_logged 'Unable to determine Go version to install: no match found for 99.99.99'
        expect(app).to_not have_logged 'Installing go99.99.99'
      end
    end

    context 'heroku example' do
      let(:app_name) { 'heroku_example/src/heroku_example' }

      specify do
        expect(app).to be_running

        browser.visit_path('/')
        expect(browser).to have_body('hello, heroku')
      end
    end

    context 'a go app using ldflags' do
      context 'with version 1.6~' do
        let(:app_name) { 'go16_ldflags/src/go_app' }

        specify do
          expect(app).to be_running
          browser.visit_path('/')
          expect(browser).to have_body('flag_linked')
          expect(app).to have_logged('main.linker_flag=flag_linked')
        end
      end
    end

    context 'app uses glide and has vendored dependencies' do
      let(:app_name) { 'glide_and_vendoring/src/glide_and_vendoring' }

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

  context 'unpackaged buildpack eg. from github' do
    let(:buildpack) { "go-unpackaged-buildpack-#{rand(1000)}" }
    let(:app) { Machete.deploy_app(app_name, buildpack: buildpack, skip_verify_version: true) }
    before do
      buildpack_file = "/tmp/#{buildpack}.zip"
      Open3.capture2e('zip','-r',buildpack_file,'bin/','src/', 'scripts/', 'manifest.yml','VERSION')[1].success? or raise 'Could not create unpackaged buildpack zip file'
      Open3.capture2e('cf', 'create-buildpack', buildpack, buildpack_file, '100', '--enable')[1].success? or raise 'Could not upload buildpack'
      FileUtils.rm buildpack_file
    end
    after do
      Open3.capture2e('cf', 'delete-buildpack', '-f', buildpack)
    end

    context 'a go app' do
      let(:app_name) { 'go_app/src/go_app' }

      it 'runs' do
        expect(app).to be_running
        expect(app).to have_logged(/Running go build supply/)
        expect(app).to have_logged(/Running go build finalize/)

        browser.visit_path('/')
        expect(browser).to have_body('go, world')
      end
    end
  end

  context 'a .godir file is detected' do
    let(:app_name) { 'deprecated_heroku/src/deprecated_heroku' }

    it 'fails with a deprecation message' do
      expect(app).to_not be_running
      expect(app).to have_logged('Deprecated, .godir file found! Please update to supported Godep or Glide dependency managers.')
      expect(app).to have_logged('See https://github.com/tools/godep or https://github.com/Masterminds/glide for usage information.')
    end
  end

  context 'a go app with wildcard matcher' do
    let(:app_name) { 'wildcard_go_version/src/go_app' }

    specify do
      expect(app).to be_running
      browser.visit_path('/')
      expect(browser).to have_body('go, world')
      expect(app).to have_logged(/Installing go 1\.6\.\d+/)
    end
  end

  context 'a go 1.6 app' do
    let(:app_name) { 'go16/src/go_app' }

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
    let(:app_name) { 'invalid_wildcard_version/src/go_app' }

    specify do
      expect(app).to_not be_running

      expect(app).to have_logged 'Unable to determine Go version to install: no match found for 1.3.x'
      expect(app).to_not have_logged 'Installing go1.3'
    end
  end
end
