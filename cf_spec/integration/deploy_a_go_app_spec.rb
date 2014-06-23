$: << 'cf_spec'
require 'spec_helper'

describe 'CF Go Buildpack' do
  subject(:app) { Machete.deploy_app(app_name) }

  context 'with cached buildpack dependencies' do
    context 'in an offline environment', if: Machete::BuildpackMode.offline? do
      context 'app has dependencies' do
        let(:app_name) { 'go_app_with_dependencies/src/go_app_with_dependencies' }

        specify do
          expect(app).to be_running
          expect(app.logs).to include 'Hello from foo!'
          expect(app.homepage_html).to include 'hello, world'
          expect(app).to have_no_internet_traffic
        end
      end

      context 'app has no dependencies' do
        let(:app_name) { 'go_app/src/go_app' }

        specify do
          expect(app).to be_running
          expect(app.homepage_html).to include 'go, world'
          expect(app).to have_no_internet_traffic
        end
      end

      context 'heroku example' do
        let(:app_name) { 'go_heroku_example/src/go_heroku_example' }

        specify do
          expect(app).to be_running
          expect(app.homepage_html).to include 'hello, heroku'
          expect(app).to have_no_internet_traffic
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
          expect(app.logs).to include 'Hello from foo!'
          expect(app.homepage_html).to include 'hello, world'
        end
      end

      context 'app has no dependencies' do
        let(:app_name) { 'go_app/src/go_app' }

        specify do
          expect(app).to be_running
          expect(app.homepage_html).to include 'go, world'
        end
      end

      context 'heroku example' do
        let(:app_name) { 'go_heroku_example/src/go_heroku_example' }

        specify do
          expect(app).to be_running
          expect(app.homepage_html).to include 'hello, heroku'
        end
      end
    end
  end
end