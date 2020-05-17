package ahhelperbot

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeliveryProvider_Unmarshal_Items(t *testing.T) {

	var jsonBytes = []byte(`{
		"_embedded": {
			"lanes": [{
				"_embedded": {
					"items": [{
						"type": "myItemType"
					}, {
						"type": "DeliveryTimeSelector",
						"_embedded": {
							"deliveryTimeSlots": [{
								"dl": 1,
								"from": "16:00",
								"state": "full",
								"to": "18:00"
							}, {
								"bdp": 77.1,
								"dl": 16,
								"from": "07:00",
								"navItem": {
									"link": {
										"href": "/kies-moment/bezorgen/1234AA/2020-04-07/1E/ODk2OA=="
									},
									"title": "order-new"
								},
								"originalValue": 7.95,
								"state": "selectable",
								"to": "08:00",
								"value": 7.95
							}]
						}
					}, {
						"type": "DeliveryDateSelector",
						"_embedded": {
							"deliveryDates": [{
								"date": "2020-04-06",
								"default": false,
								"deliveryTimeSlots": [{
									"bdp": 100.0,
									"dl": 0,
									"from": "16:00",
									"state": "full",
									"to": "18:00"
								}]
							}]
						}
					}]
				}
			}]
		}
	}`)

	dr := deliveryResponse{}

	// Act
	err := json.Unmarshal(jsonBytes, &dr)

	assert.NoError(t, err)
	assert.NotEmpty(t, dr.lanes)
	assert.NotEmpty(t, dr.lanes[0].items)

	assert.NotEmpty(t, dr.lanes[0].items[1].deliveryTimeSlots)
	assert.Equal(t, 1, dr.lanes[0].items[1].deliveryTimeSlots[0].Dl)
	assert.Equal(t, "16:00", dr.lanes[0].items[1].deliveryTimeSlots[0].From)
	assert.Equal(t, "full", dr.lanes[0].items[1].deliveryTimeSlots[0].State)
	assert.Equal(t, "18:00", dr.lanes[0].items[1].deliveryTimeSlots[0].To)

	assert.Equal(t, 16, dr.lanes[0].items[1].deliveryTimeSlots[1].Dl)
	assert.Equal(t, "07:00", dr.lanes[0].items[1].deliveryTimeSlots[1].From)
	assert.Equal(t, "selectable", dr.lanes[0].items[1].deliveryTimeSlots[1].State)
	assert.Equal(t, "08:00", dr.lanes[0].items[1].deliveryTimeSlots[1].To)
	assert.Equal(t, "2020-04-07", dr.lanes[0].items[1].deliveryTimeSlots[1].Date)

	assert.NotEmpty(t, dr.lanes[0].items[2].deliveryDates)
	assert.Equal(t, "2020-04-06", dr.lanes[0].items[2].deliveryDates[0].Date)
	assert.NotEmpty(t, dr.lanes[0].items[2].deliveryDates[0].DeliveryTimeSlots)
	assert.Equal(t, "16:00", dr.lanes[0].items[2].deliveryDates[0].DeliveryTimeSlots[0].From)
	assert.Equal(t, "full", dr.lanes[0].items[2].deliveryDates[0].DeliveryTimeSlots[0].State)
	assert.Equal(t, "18:00", dr.lanes[0].items[2].deliveryDates[0].DeliveryTimeSlots[0].To)
}

func TestDeliveryProvider_convertResponseToSchedule(t *testing.T) {

	dr := deliveryResponse{
		[]deliveryLane{
			deliveryLane{
				items: []item{
					item{
						deliveryDates: []deliveryDate{
							deliveryDate{
								Date: "d1-1",
								DeliveryTimeSlots: []deliveryTimeSlot{
									deliveryTimeSlot{
										Date: "d1-1",
										DeliveryTimeSlotBase: DeliveryTimeSlotBase{
											Dl:    1,
											From:  "f1-1",
											To:    "t1-1",
											State: "s1-1",
										},
									},
									deliveryTimeSlot{
										Date: "d1-1",
										DeliveryTimeSlotBase: DeliveryTimeSlotBase{
											Dl:    2,
											From:  "f1-2",
											To:    "t1-2",
											State: "full",
										},
									},
									deliveryTimeSlot{
										Date: "d1-1",
										DeliveryTimeSlotBase: DeliveryTimeSlotBase{
											Dl:    3,
											From:  "f1-3",
											To:    "t1-3",
											State: "s1-3",
										},
									},
								},
							},
						},
						deliveryTimeSlots: []deliveryTimeSlot{
							deliveryTimeSlot{
								Date: "d2-1",
								DeliveryTimeSlotBase: DeliveryTimeSlotBase{
									Dl:    4,
									From:  "f2-1",
									To:    "t2-1",
									State: "s2-1",
								},
							},
							deliveryTimeSlot{
								Date: "d2-2",
								DeliveryTimeSlotBase: DeliveryTimeSlotBase{
									Dl:    5,
									From:  "f2-2",
									To:    "t2-2",
									State: "full",
								},
							},
							deliveryTimeSlot{
								Date: "d2-3",
								DeliveryTimeSlotBase: DeliveryTimeSlotBase{
									Dl:    5,
									From:  "f2-3",
									To:    "t2-3",
									State: "s2-3",
								},
							},
						},
					},
				},
			},
		},
	}

	// Act
	ds := convertResponseToSchedule(dr)

	assert.NotNil(t, ds)
	assert.Equal(t, 2, len(ds["d1-1"]))
	assert.Equal(t, "f1-1", ds["d1-1"][0].From)
	assert.Equal(t, "t1-1", ds["d1-1"][0].To)
	assert.Equal(t, "f1-3", ds["d1-1"][1].From)
	assert.Equal(t, "t1-3", ds["d1-1"][1].To)

	assert.Equal(t, "f2-1", ds["d2-1"][0].From)
	assert.Equal(t, "t2-1", ds["d2-1"][0].To)
	assert.Equal(t, "f2-3", ds["d2-3"][0].From)
	assert.Equal(t, "t2-3", ds["d2-3"][0].To)
}
