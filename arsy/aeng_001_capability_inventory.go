package aeng

// CapabilityInventory records PocketBase record and collection capabilities.
// Auto-generated from AENG-001 design documentation.
type CapabilityInventory struct {
	Collections map[string]CollectionCapability
}

// CollectionCapability holds per-collection metadata.
type CollectionCapability struct {
	Name      string
	Fields    map[string]FieldCapability
	Indexes   []string
	Relations []RelationCapability
	APIRules  map[string]map[string]bool // action -> allowed roles
}

// FieldCapability documents a PocketBase field.
type FieldCapability struct {
	Name           string
	PocketBaseType string
	GoType         string
	Required       bool
	Unique         bool
	Max            int
	Pattern        string
}

// RelationCapability documents a PocketBase relation.
type RelationCapability struct {
	Name       string
	Target     string // target collection/go type
	Type       string // "has_one", "has_many", "many_to_many"
	ForeignKey string // FK field on child
	OwnerKey   string // PK field on parent
}
