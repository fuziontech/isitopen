require "uri"
require "net/http"

def getRoadStatus(road)
  statusPatternMap = {
    "NO TRAFFIC RESTRICTIONS ARE REPORTED FOR THIS AREA." => "OPEN",
    "CHAINS ARE REQUIRED " => "CHAINS",
    "ADVISORY" => "ADVISORY",
    "CLOSED" => "CLOSED",
    "CONSTRUCTION" => "CONSTRUCTION"
  }
  params = {
    "roadnumber" => road,
  }
  resp = Net::HTTP.post_form(URI.parse('https://roads.dot.ca.gov/'), params)
  doc = Nokogiri::HTML5.parse(resp.body)
  statusDoc = doc.search(".main-primary p").text
  statusSplit = statusDoc.tr("/", "").split("\n")
  statusStringTime = statusSplit[0]
  statusTime = DateTime.parse(statusStringTime.split(",").drop(1).join("").tr('.', ''))
  statusText = statusSplit.drop(1).join(" ")
  status = "OPEN"
  statusPatternMap.keys.each do |pattern|
    if statusText.include? pattern
      status = statusPatternMap[pattern]
    end
  end
  return statusTime, status, statusText
end

class StatusController < ApplicationController
  def index
    hostLookup = {
      "is50open.com" => 50,
      "is80open.com" => 80,
      "is88open.com" => 88,
    }
    if hostLookup.include?(request.host)
      road = hostLookup[request.host]
    elsif params[:road]
      road = params[:road]
    else
      road = 50
    end
    statusTime, status, statusText = getRoadStatus(road)
    @status = status
    @description = statusText
    @statusTime = statusTime.strftime("%B %d, %Y at %-I:%M%p")
    @name = road
  end

  def api
    road = params[:road]
    statusTime, status, statusText = getRoadStatus(road)
    statusMap = {
      "name" => road,
      "status" => status,
      "description" => statusText,
      "UpdatedAt" => statusTime,
    }
    render json: statusMap
  end
end
