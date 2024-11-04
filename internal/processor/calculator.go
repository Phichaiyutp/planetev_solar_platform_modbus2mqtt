package processor

import (
	"bytes"
	"encoding/binary"
	"log"

	"modbus-mqtt-service/internal/models"
	"modbus-mqtt-service/pkg/utils"
)

func processRegisterData(group []models.Register, data []byte, payload *models.SensorEnergy) {
	dataIndex := 0

	for _, reg := range group {
		var dataSize int
		switch reg.Type {
		case "float32_be":
			dataSize = 4
		case "int16":
			dataSize = 2
		default:
			log.Printf("Unknown data type: %s", reg.Type)
			continue
		}

		if dataIndex+dataSize > len(data) {
			log.Printf("Data length mismatch while processing register %s", reg.Name)
			continue
		}

		dataChunk := data[dataIndex : dataIndex+dataSize]
		dataIndex += dataSize

		value := decodeValue(reg, dataChunk)
		payload.Data = append(payload.Data, models.DataEntry{
			Name:  reg.Name,
			Value: value,
			Unit:  reg.Unit,
		})
	}

	calculateAdditionalData(payload)
}

func decodeValue(reg models.Register, data []byte) interface{} {
	switch reg.Type {
	case "float32_be":
		var floatValue float32
		buf := bytes.NewReader(data)
		if err := binary.Read(buf, binary.BigEndian, &floatValue); err != nil {
			return nil
		}
		return utils.ToFixed(float64(floatValue)*reg.Ratio, 3)

	case "int16":
		var intValue int16
		buf := bytes.NewReader(data)
		if err := binary.Read(buf, binary.BigEndian, &intValue); err != nil {
			return nil
		}
		return utils.ToFixed(float64(intValue)*reg.Ratio, 3)
	}
	return nil
}

// calculateAdditionalData performs custom calculations and adds them to the payload.
func calculateAdditionalData(payload *models.SensorEnergy) {
	// Create a map to store the extracted values for calculations
	values := make(map[string]float64)

	// Create a helper function to find an index of an existing data entry
	findDataIndex := func(name string) int {
		for index, item := range payload.Data {
			if item.Name == name {
				return index
			}
		}
		return -1
	}

	// Extract values for calculations if they exist
	for _, item := range payload.Data {
		if v, ok := item.Value.(float64); ok {
			values[item.Name] = v
		}
	}

	// Calculate U_RMS (Voltage)
	if ua, uaOk := values["Ua"]; uaOk {
		if ub, ubOk := values["Ub"]; ubOk {
			if uc, ucOk := values["Uc"]; ucOk {
				if urAt, urAtOk := values["UrAt"]; urAtOk {
					urms := (ua + ub + uc) / 3
					uValue := utils.ToFixed(urms*urAt, 3)

					// Update or add the U value
					if index := findDataIndex("U"); index != -1 {
						payload.Data[index].Value = uValue
					} else {
						payload.Data = append(payload.Data, models.DataEntry{
							Name:  "U",
							Value: uValue,
							Unit:  "V",
						})
					}
				}
			}
		}
	}

	// Calculate I_RMS (Current)
	if ia, iaOk := values["Ia"]; iaOk {
		if ib, ibOk := values["Ib"]; ibOk {
			if ic, icOk := values["Ic"]; icOk {
				if irAt, irAtOk := values["IrAt"]; irAtOk {
					irms := (ia + ib + ic) / 3
					iValue := utils.ToFixed(irms*irAt, 3)

					// Update or add the I value
					if index := findDataIndex("I"); index != -1 {
						payload.Data[index].Value = iValue
					} else {
						payload.Data = append(payload.Data, models.DataEntry{
							Name:  "I",
							Value: iValue,
							Unit:  "A",
						})
					}
				}
			}
		}
	}

	// Calculate Power
	if pt, ptOk := values["Pt"]; ptOk {
		if urAt, urAtOk := values["UrAt"]; urAtOk {
			if irAt, irAtOk := values["IrAt"]; irAtOk {
				power := utils.ToFixed(pt*urAt*irAt, 3)

				// Update or add the P value
				if index := findDataIndex("P"); index != -1 {
					payload.Data[index].Value = power
				} else {
					payload.Data = append(payload.Data, models.DataEntry{
						Name:  "P",
						Value: power,
						Unit:  "W",
					})
				}
			}
		}
	}
}
