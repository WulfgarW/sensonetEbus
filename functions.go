package sensonetEbus

func GetZoneData(zones []VaillantRelDataZones, index int) *VaillantRelDataZones {
	// Extracting correct Zones element
	if len(zones) == 0 {
		return nil
	}
	for _, zone := range zones {
		if zone.Index == index || (zone.Index == ZONEINDEX_DEFAULT && index < 0) {
			return &zone
		}
	}
	return nil
}
