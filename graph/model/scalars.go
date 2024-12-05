package model

import (
	"encoding/json"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/pkg/errors"
	"io"
)

type GapToLeader struct {
	Value interface{}
}

func (g *GapToLeader) UnmarshalJSON(data []byte) error {
	var err error
	var floatValue float64
	var stringValue string

	fmt.Println("over here")

	if err = json.Unmarshal(data, &floatValue); err == nil {
		g.Value = floatValue
		return nil
	}

	if err = json.Unmarshal(data, &stringValue); err == nil {
		g.Value = stringValue
		return nil
	}

	return fmt.Errorf("unexpected type for GapToLeader: %s", data)
}

func (g GapToLeader) MarshalJSON() ([]byte, error) {
	return json.Marshal(g.Value)
}

func MarshalGapToLeader(g GapToLeader) graphql.Marshaler {
	fmt.Println("there")
	return graphql.WriterFunc(func(w io.Writer) {
		switch v := g.Value.(type) {
		case float64:
			fmt.Fprintf(w, "%f", v)
		case string:
			fmt.Fprintf(w, "%q", v)
		default:
			fmt.Fprintf(w, "%q", "")
		}
	})
}

func UnmarshalGapToLeader(v interface{}) (GapToLeader, error) {
	fmt.Println("here")
	switch v := v.(type) {
	case float64:
		return GapToLeader{Value: v}, nil
	case string:
		return GapToLeader{Value: v}, nil
	default:
		return GapToLeader{}, errors.New("invalid type for GapToLeader")
	}
}
