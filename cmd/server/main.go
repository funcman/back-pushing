package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"

	graphql "github.com/graphql-go/graphql"

	"github.com/funcman/back-pushing/internal/action"
	"github.com/funcman/back-pushing/internal/engine/graph"
	"github.com/funcman/back-pushing/internal/ontology"
	"github.com/funcman/back-pushing/internal/storage/memory"
)

// Server is the main server struct that wires all components together
type Server struct {
	ontology       *ontology.Ontology
	objectStore   *memory.ObjectStore
	graphStore    *memory.GraphStore
	traversal     *graph.TraversalEngine
	search        *graph.SearchEngine
	dispatcher    *action.Dispatcher
	actionContext *action.ActionContext
	mu            sync.RWMutex
}

func NewServer(ont *ontology.Ontology) *Server {
	objStore := memory.NewObjectStore()
	grpStore := memory.NewGraphStore()
	traversal := graph.NewTraversalEngine(grpStore)
	search := graph.NewSearchEngine(traversal, objStore)
	dispatcher := action.NewDispatcher()
	actionCtx := action.NewActionContext(objStore)

	return &Server{
		ontology:       ont,
		objectStore:    objStore,
		graphStore:     grpStore,
		traversal:      traversal,
		search:         search,
		dispatcher:     dispatcher,
		actionContext:  actionCtx,
	}
}

func (s *Server) RegisterActions() {
	s.dispatcher.Register("action.escalate_review", action.EscalateReview)
}

func (s *Server) RegisterLinks() {
	for name := range s.ontology.Links {
		s.traversal.RegisterLink(name, name)
	}
}

// MapScalar is a custom GraphQL scalar for map[string]any (JSON object)
var MapScalar = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "Map",
	Description: "JSON object represented as map[string]any",
	Serialize: func(value interface{}) interface{} {
		if m, ok := value.(map[string]any); ok {
			return m
		}
		if m, ok := value.(map[string]interface{}); ok {
			result := make(map[string]any)
			for k, v := range m {
				result[k] = v
			}
			return result
		}
		return nil
	},
	ParseValue: func(value interface{}) interface{} {
		if m, ok := value.(map[string]any); ok {
			return m
		}
		if m, ok := value.(map[string]interface{}); ok {
			result := make(map[string]any)
			for k, v := range m {
				result[k] = v
			}
			return result
		}
		return nil
	},
})

// ObjectType defines the GraphQL Object type
var objectType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Object",
	Fields: graphql.Fields{
		"type": &graphql.Field{Type: graphql.String},
		"id":   &graphql.Field{Type: graphql.String},
		"data": &graphql.Field{Type: MapScalar},
	},
})

// SearchResultType defines the GraphQL SearchResult type
var searchResultType = graphql.NewObject(graphql.ObjectConfig{
	Name: "SearchResult",
	Fields: graphql.Fields{
		"nodeId": &graphql.Field{Type: graphql.String},
		"type":   &graphql.Field{Type: graphql.String},
		"data":   &graphql.Field{Type: MapScalar},
	},
})

// ActionResultType defines the GraphQL ActionResult type
var actionResultType = graphql.NewObject(graphql.ObjectConfig{
	Name: "ActionResult",
	Fields: graphql.Fields{
		"caseId":     &graphql.Field{Type: graphql.String},
		"assignedTo": &graphql.Field{Type: graphql.String},
		"status":     &graphql.Field{Type: graphql.String},
		"timestamp":  &graphql.Field{Type: graphql.String},
	},
})

// EdgeType defines the GraphQL Edge type
var edgeType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Edge",
	Fields: graphql.Fields{
		"from":  &graphql.Field{Type: graphql.String},
		"to":    &graphql.Field{Type: graphql.String},
		"props": &graphql.Field{Type: MapScalar},
	},
})

func (s *Server) handleGraphQL(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Query         string                 `json:"query"`
		OperationName string                `json:"operationName"`
		Variables     map[string]any        `json:"variables"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	result := graphql.Do(graphql.Params{
		Schema:         s.schema(),
		RequestString:  request.Query,
		VariableValues: request.Variables,
		OperationName:  request.OperationName,
		Context:        r.Context(),
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) schema() graphql.Schema {
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"object": &graphql.Field{
				Type: objectType,
				Args: graphql.FieldConfigArgument{
					"type": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"id":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					objType, _ := p.Args["type"].(string)
					id, _ := p.Args["id"].(string)
					ctx := context.Background()
					data, err := s.objectStore.Get(ctx, objType, id)
					if err != nil {
						return nil, err
					}
					return map[string]any{
						"type": objType,
						"id":   id,
						"data": data,
					}, nil
				},
			},
			"search": &graphql.Field{
				Type: graphql.NewList(searchResultType),
				Args: graphql.FieldConfigArgument{
					"query": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"type":  &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					query, _ := p.Args["query"].(string)
					objType, _ := p.Args["type"].(string)
					ctx := context.Background()
					results, err := s.search.FullTextSearch(ctx, objType, query)
					if err != nil {
						return nil, err
					}
					var output []map[string]any
					for _, r := range results {
						output = append(output, map[string]any{
							"nodeId": r.NodeID,
							"type":   r.Type,
							"data":   r.Data,
						})
					}
					return output, nil
				},
			},
			"edges": &graphql.Field{
				Type: graphql.NewList(edgeType),
				Args: graphql.FieldConfigArgument{
					"linkType": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"nodeId":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					linkType, _ := p.Args["linkType"].(string)
					nodeID, _ := p.Args["nodeId"].(string)
					ctx := context.Background()
					edges, err := s.graphStore.GetEdges(ctx, linkType, nodeID)
					if err != nil {
						return nil, err
					}
					var output []map[string]any
					for _, e := range edges {
						output = append(output, map[string]any{
							"from":  e.From,
							"to":    e.To,
							"props": e.Props,
						})
					}
					return output, nil
				},
			},
		},
	})

	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createObject": &graphql.Field{
				Type: objectType,
				Args: graphql.FieldConfigArgument{
					"type": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"id":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"data": &graphql.ArgumentConfig{Type: graphql.NewNonNull(MapScalar)},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					objType, _ := p.Args["type"].(string)
					id, _ := p.Args["id"].(string)
					data, _ := p.Args["data"].(map[string]any)
					ctx := context.Background()
					err := s.objectStore.Create(ctx, objType, id, data)
					if err != nil {
						return nil, err
					}
					return map[string]any{
						"type": objType,
						"id":   id,
						"data": data,
					}, nil
				},
			},
			"addEdge": &graphql.Field{
				Type: graphql.Boolean,
				Args: graphql.FieldConfigArgument{
					"linkType": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"fromId":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"toId":     &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"props":    &graphql.ArgumentConfig{Type: MapScalar},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					linkType, _ := p.Args["linkType"].(string)
					fromID, _ := p.Args["fromId"].(string)
					toID, _ := p.Args["toId"].(string)
					props, _ := p.Args["props"].(map[string]any)
					if props == nil {
						props = make(map[string]any)
					}
					ctx := context.Background()
					err := s.graphStore.AddEdge(ctx, linkType, fromID, toID, props)
					return err == nil, err
				},
			},
			"invokeAction": &graphql.Field{
				Type: actionResultType,
				Args: graphql.FieldConfigArgument{
					"name":  &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"input": &graphql.ArgumentConfig{Type: graphql.NewNonNull(MapScalar)},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					name, _ := p.Args["name"].(string)
					input, _ := p.Args["input"].(map[string]any)
					result, err := s.dispatcher.Dispatch(s.actionContext, name, input)
					if err != nil {
						return nil, err
					}
					return result, nil
				},
			},
		},
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	})
	if err != nil {
		panic(err)
	}
	return schema
}

func main() {
	ontologyDir := flag.String("ontology", "./ontology", "path to ontology directory")
	addr := flag.String("addr", ":8080", "server address")
	flag.Parse()

	ont, err := ontology.ParseOntology(*ontologyDir)
	if err != nil {
		log.Fatalf("failed to load ontology: %v", err)
	}

	srv := NewServer(ont)
	srv.RegisterActions()
	srv.RegisterLinks()

	http.HandleFunc("/graphql", srv.handleGraphQL)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `<html>
			<head><title>Back-Pushing Server</title></head>
			<body>
				<h1>Back-Pushing GraphQL Server</h1>
				<p>GraphQL endpoint: <a href="/graphql">/graphql</a></p>
				<p>Send POST requests to /graphql with JSON body:</p>
				<pre>{"query":"{ object(type: \"Person\", id: \"1\") { type id data } }"}</pre>
			</body>
		</html>`)
	})

	log.Printf("server starting on %s", *addr)
	log.Printf("GraphQL endpoint available at http://localhost%s/graphql", *addr)

	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
