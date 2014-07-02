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
          expect(app).to have_logged('Hello from foo!')
          expect(app).to have_page_body('hello, world')
          expect(app.host).not_to have_internet_traffic
        end
      end

      context 'app has no dependencies' do
        let(:app_name) { 'go_app/src/go_app' }

        specify do
          expect(app).to be_running
          expect(app).to have_page_body('go, world')
          expect(app.host).not_to have_internet_traffic
        end
      end

      context 'heroku example' do
        let(:app_name) { 'go_heroku_example/src/go_heroku_example' }

        specify do
          expect(app).to be_running
          expect(app).to have_page_body('hello, heroku')
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
          expect(app).to have_page_body('hello, world')
        end
      end

      context 'app has no dependencies' do
        let(:app_name) { 'go_app/src/go_app' }

        specify do
          expect(app).to be_running
          expect(app).to have_page_body('go, world')
        end
      end

      context 'heroku example' do
        let(:app_name) { 'go_heroku_example/src/go_heroku_example' }

        specify do
          expect(app).to be_running
          expect(app).to have_page_body('hello, heroku')
        end
      end
    end
  end
end