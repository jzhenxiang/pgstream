package wal

import (
	"fmt"

	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5/pgtype"
)

// Decoder decodes raw WAL messages into structured Events.
type Decoder struct {
	typeMap  *pgtype.Map
	relations map[uint32]*pglogrepl.RelationMessageV2
}

// NewDecoder creates a new Decoder.
func NewDecoder() *Decoder {
	return &Decoder{
		typeMap:   pgtype.NewMap(),
		relations: make(map[uint32]*pglogrepl.RelationMessageV2),
	}
}

// Decode parses a raw Message and returns an Event, or nil for non-data messages.
func (d *Decoder) Decode(msg Message) (*Event, error) {
	if len(msg.Data) == 0 {
		return nil, nil
	}

	// Skip keepalive byte (0x77 = 'w')
	if msg.Data[0] != pglogrepl.XLogDataByteID {
		return nil, nil
	}

	xld, err := pglogrepl.ParseXLogData(msg.Data[1:])
	if err != nil {
		return nil, fmt.Errorf("decoder: parse xlog: %w", err)
	}

	logicalMsg, err := pglogrepl.ParseV2(xld.WALData, false)
	if err != nil {
		return nil, fmt.Errorf("decoder: parse logical: %w", err)
	}

	lsn := xld.WALStart.String()

	switch m := logicalMsg.(type) {
	case *pglogrepl.RelationMessageV2:
		d.relations[m.RelationID] = m
		return nil, nil

	case *pglogrepl.InsertMessageV2:
		rel, ok := d.relations[m.RelationID]
		if !ok {
			return nil, fmt.Errorf("decoder: unknown relation %d", m.RelationID)
		}
		cols, err := d.decodeColumns(rel, m.Tuple)
		if err != nil {
			return nil, err
		}
		return &Event{Type: MessageTypeInsert, LSN: lsn, Schema: rel.Namespace, Table: rel.RelationName, Columns: cols}, nil

	case *pglogrepl.UpdateMessageV2:
		rel, ok := d.relations[m.RelationID]
		if !ok {
			return nil, fmt.Errorf("decoder: unknown relation %d", m.RelationID)
		}
		newCols, err := d.decodeColumns(rel, m.NewTuple)
		if err != nil {
			return nil, err
		}
		return &Event{Type: MessageTypeUpdate, LSN: lsn, Schema: rel.Namespace, Table: rel.RelationName, Columns: newCols}, nil

	case *pglogrepl.DeleteMessageV2:
		rel, ok := d.relations[m.RelationID]
		if !ok {
			return nil, fmt.Errorf("decoder: unknown relation %d", m.RelationID)
		}
		oldCols, err := d.decodeColumns(rel, m.OldTuple)
		if err != nil {
			return nil, err
		}
		return &Event{Type: MessageTypeDelete, LSN: lsn, Schema: rel.Namespace, Table: rel.RelationName, OldColumns: oldCols}, nil
	}

	return nil, nil
}

func (d *Decoder) decodeColumns(rel *pglogrepl.RelationMessageV2, tuple *pglogrepl.TupleData) (map[string]any, error) {
	if tuple == nil {
		return nil, nil
	}
	result := make(map[string]any, len(tuple.Columns))
	for i, col := range tuple.Columns {
		if i >= len(rel.Columns) {
			break
		}
		name := rel.Columns[i].Name
		switch col.DataType {
		case 'n':
			result[name] = nil
		case 't':
			result[name] = string(col.Data)
		}
	}
	return result, nil
}
