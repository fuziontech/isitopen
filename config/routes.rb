Rails.application.routes.draw do
  get 'status', to: 'status#index'
  get 'road/:road', to: 'status#index'

  scope '/v1' do
    get 'road/:road', to: 'status#api'
  end
  # For details on the DSL available within this file, see https://guides.rubyonrails.org/routing.html

  # Almost every application defines a route for the root path ("/") at the top of this file.
  root "status#index"
end
