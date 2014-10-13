$: << 'cf_spec'
require 'spec_helper'

describe 'CF Go Buildpack' do
  subject(:app) { Machete.deploy_app(app_name) }
  let(:browser) { Machete::Browser.new(app) }

  context 'with cached buildpack dependencies' do
    context 'in an offline environment', if: Machete::BuildpackMode.offline? do
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

      context 'app has no dependencies' do
        let(:app_name) { 'go_app/src/go_app' }

        specify do
          expect(app).to be_running

          browser.visit_path('/')
          expect(browser).to have_body('go, world')

          expect(app.host).not_to have_internet_traffic
        end
      end

      context 'expects a non-packaged version of go' do
        let(:app_name) { 'go99/src/go99' }
        let(:resource_url) { "https://storage.googleapis.com/golang/go99.99.99.linux-amd64.tar.gz" }

        it "displays useful understandable errors" do
          expect(app).not_to be_running

          expect(app).to have_logged("Resource #{resource_url} is not provided by this buildpack. Please upgrade your buildpack to receive the latest resources.")

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
    end
  end

  context 'without cached buildpack dependencies' do
    context 'in an online environment', if: Machete::BuildpackMode.online? do
      context 'app has dependencies' do
        let(:app_name) { 'go_app_with_dependencies/src/go_app_with_dependencies' }

        specify do
          expect(app).to be_running
          expect(app).to have_logged('Hello from foo!')

          browser.visit_path('/')
          expect(browser).to have_body('hello, world')
        end
      end

      context 'app has no dependencies' do
        let(:app_name) { 'go_app/src/go_app' }

        specify do
          expect(app).to be_running

          browser.visit_path('/')
          expect(browser).to have_body('go, world')
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
    end
  end
end
