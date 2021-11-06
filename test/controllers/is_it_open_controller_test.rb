require "test_helper"

class IsItOpenControllerTest < ActionDispatch::IntegrationTest
  test "should get status" do
    get is_it_open_status_url
    assert_response :success
  end
end
