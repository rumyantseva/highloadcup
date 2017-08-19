package db

import memdb "github.com/hashicorp/go-memdb"

// Schema ...
func Schema() *memdb.DBSchema {
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"user": &memdb.TableSchema{
				Name: "user",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.UintFieldIndex{Field: "ID"},
					},
				},
			},

			"location": &memdb.TableSchema{
				Name: "location",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.UintFieldIndex{Field: "ID"},
					},
				},
			},

			"visit": &memdb.TableSchema{
				Name: "visit",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.UintFieldIndex{Field: "ID"},
					},
					"user_id": &memdb.IndexSchema{
						Name:    "user_id",
						Unique:  false,
						Indexer: &memdb.UintFieldIndex{Field: "User"},
					},
					"location_id": &memdb.IndexSchema{
						Name:    "location_id",
						Unique:  false,
						Indexer: &memdb.UintFieldIndex{Field: "Location"},
					},
				},
			},
		},
	}

	return schema
}
