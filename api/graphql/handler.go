package graphql

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/graphql-go/graphql"
    "github.com/graphql-go/graphql/language/ast"

    "github.com/funcman/back-pushing/internal/storage/memory"
)

var JSONScalar = graphql.NewScalar(graphql.ScalarConfig{
    Name:        "JSON",
    Description: "The JSON scalar type represents arbitrary JSON value",
    Serialize: func(value interface{}) interface{} {
        return value
    },
    ParseValue: func(value interface{}) interface{} {
        return value
    },
    ParseLiteral: func(valueAST ast.Value) interface{} {
        return valueAST
    },
})

func NewHandler(store *memory.ObjectStore) gin.HandlerFunc {
    resolver := NewResolver(store)

    queryType := graphql.NewObject(graphql.ObjectConfig{
        Name: "Query",
        Fields: graphql.Fields{
            "object": &graphql.Field{
                Type: graphql.NewObject(graphql.ObjectConfig{
                    Name: "Object",
                    Fields: graphql.Fields{
                        "id":   &graphql.Field{Type: graphql.String},
                        "type": &graphql.Field{Type: graphql.String},
                        "data": &graphql.Field{Type: JSONScalar},
                    },
                }),
                Args: graphql.FieldConfigArgument{
                    "type": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
                    "id":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
                },
                Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                    return resolver.Object(p.Context, struct{ Type, ID string }{
                        Type: p.Args["type"].(string),
                        ID:   p.Args["id"].(string),
                    })
                },
            },
            "objects": &graphql.Field{
                Type: graphql.NewList(graphql.NewObject(graphql.ObjectConfig{
                    Name: "Object",
                    Fields: graphql.Fields{
                        "id":   &graphql.Field{Type: graphql.String},
                        "type": &graphql.Field{Type: graphql.String},
                        "data": &graphql.Field{Type: JSONScalar},
                    },
                })),
                Args: graphql.FieldConfigArgument{
                    "type": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
                },
                Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                    return resolver.Objects(p.Context, struct{ Type string }{Type: p.Args["type"].(string)})
                },
            },
        },
    })

    mutationType := graphql.NewObject(graphql.ObjectConfig{
        Name: "Mutation",
        Fields: graphql.Fields{
            "createObject": &graphql.Field{
                Type: graphql.NewObject(graphql.ObjectConfig{
                    Name: "Object",
                    Fields: graphql.Fields{
                        "id":   &graphql.Field{Type: graphql.String},
                        "type": &graphql.Field{Type: graphql.String},
                        "data": &graphql.Field{Type: JSONScalar},
                    },
                }),
                Args: graphql.FieldConfigArgument{
                    "type": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
                    "id":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
                    "data": &graphql.ArgumentConfig{Type: graphql.NewNonNull(JSONScalar)},
                },
                Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                    return resolver.CreateObject(p.Context, struct {
                        Type string         `json:"type"`
                        ID   string         `json:"id"`
                        Data map[string]any `json:"data"`
                    }{
                        Type: p.Args["type"].(string),
                        ID:   p.Args["id"].(string),
                        Data: p.Args["data"].(map[string]any),
                    })
                },
            },
        },
    })

    schema, _ := graphql.NewSchema(graphql.SchemaConfig{
        Query:    queryType,
        Mutation: mutationType,
    })

    return func(c *gin.Context) {
        var req struct {
            Query         string                 `json:"query"`
            OperationName string                 `json:"operationName"`
            Variables     map[string]interface{} `json:"variables"`
        }
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        result := graphql.Do(graphql.Params{
            Schema:         schema,
            RequestString:  req.Query,
            VariableValues: req.Variables,
            OperationName:  req.OperationName,
        })
        c.JSON(http.StatusOK, result)
    }
}