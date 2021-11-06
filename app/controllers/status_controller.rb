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
  # lookup cache
  cachedStatus = RoadStatus.where('"roadName" = ? AND "calTransUpdatedAt" >= ?', road.to_s, (Time.current - (10 * 60 * 60))).first
  if cachedStatus
    return cachedStatus
  end

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
  cachedStatus = RoadStatus.new
  cachedStatus.roadName = road
  cachedStatus.status = status
  cachedStatus.calTransUpdatedAt = statusTime
  cachedStatus.save
  return cachedStatus
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
    status = getRoadStatus(road)
    @status = status.status
    @description = status.description
    @statusTime = status.calTransUpdatedAt.strftime("%B %d, %Y at %-I:%M%p")
    @name = status.roadName
  end

  def api
    road = params[:road]
    status = getRoadStatus(road)
    statusMap = {
      "name" => status.roadName,
      "status" => status.status,
      "description" => status.description,
      "UpdatedAt" => status.calTransUpdatedAt,
    }
    render json: statusMap
  end
end
