package linnworks

import (
	"net/http"

	"polymetrics.ai/internal/connectors"
)

type streamEndpoint struct {
	resource    string
	method      string
	recordsPath string
	mapRecord   func(map[string]any) connectors.Record
}

var streamEndpoints = map[string]streamEndpoint{
	"inventory_items": {resource: "Inventory/GetInventoryItems", method: http.MethodPost, recordsPath: "Data", mapRecord: inventoryRecord},
	"open_orders":     {resource: "Orders/GetOpenOrders", method: http.MethodPost, recordsPath: "Data", mapRecord: orderRecord},
	"warehouses":      {resource: "Warehouses/GetWarehouses", method: http.MethodPost, recordsPath: "Data", mapRecord: warehouseRecord},
	"suppliers":       {resource: "Inventory/GetSuppliers", method: http.MethodPost, recordsPath: "Data", mapRecord: supplierRecord},
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "inventory_items", Description: "Linnworks inventory items.", PrimaryKey: []string{"sku"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "sku", Type: "string"}, {Name: "title", Type: "string"}, {Name: "quantity", Type: "number"}}},
		{Name: "open_orders", Description: "Linnworks open orders.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "reference", Type: "string"}, {Name: "status", Type: "string"}}},
		{Name: "warehouses", Description: "Linnworks warehouses.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}}},
		{Name: "suppliers", Description: "Linnworks suppliers.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}}},
	}
}

func inventoryRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "pkStockItemId", "Id", "id"), "sku": first(item, "ItemNumber", "SKU", "sku"), "title": first(item, "ItemTitle", "Title", "title"), "quantity": first(item, "Quantity", "quantity")}
}
func orderRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "OrderId", "pkOrderID", "id"), "reference": first(item, "ReferenceNum", "reference"), "status": first(item, "Status", "status")}
}
func warehouseRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "pkStockLocationId", "Id", "id"), "name": first(item, "Location", "Name", "name")}
}
func supplierRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "pkSupplierID", "Id", "id"), "name": first(item, "SupplierName", "Name", "name")}
}
func first(item map[string]any, keys ...string) any {
	for _, key := range keys {
		if item[key] != nil {
			return item[key]
		}
	}
	return nil
}
