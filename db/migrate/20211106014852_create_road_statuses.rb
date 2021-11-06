class CreateRoadStatuses < ActiveRecord::Migration[7.0]
  def change
    create_table :road_statuses do |t|
      t.text :roadName
      t.text :status
      t.text :description
      t.datetime :calTransUpdatedAt

      t.timestamps
    end
  end
end
