require "test_helper"

class StatusControllerTest < ActionDispatch::IntegrationTest
  test "should get status" do
    get status_status_url
    assert_response :success
  end
end
