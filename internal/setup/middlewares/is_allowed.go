package middlewares

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Member struct {
	ID       primitive.ObjectID `bson:"_id"`
	Role     string             `bson:"role"`
	MemberId primitive.ObjectID `bson:"memberId"`
}

type Workspace struct {
	ID      primitive.ObjectID `bson:"_id"`
	Owner   primitive.ObjectID `bson:"owner"`
	Members []Member           `bson:"members"`
}

func IsAllowed(next http.Handler, db *mongo.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		workspaceId := r.Header.Get("workspaceId")
		// applicationId := r.Header.Get("applicationId")

		collection := db.Collection("workspaces")
		workspaceObjectID, err := primitive.ObjectIDFromHex(workspaceId)
		if err != nil {
			http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
			return
		}

		result := collection.FindOne(helpers.Ctx, bson.M{"_id": workspaceObjectID})

		if result.Err() != nil {
			http.Error(w, "Workspace not found", http.StatusNotFound)
			return
		}

		var workspace Workspace
		if err := result.Decode(&workspace); err != nil {
			http.Error(w, "Error decoding workspace", http.StatusInternalServerError)
			return
		}

		userId := r.Header.Get("UserId")
		userObjectID, err := primitive.ObjectIDFromHex(userId)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		isUserAllowed := false

		if workspace.Owner == userObjectID {
			isUserAllowed = true
		} else {
			for _, value := range workspace.Members {
				if value.MemberId == userObjectID && value.Role == "admin" {
					isUserAllowed = true
					break
				}
			}
		}

		// tenho que fazer com que ele verifique se esse user tem acesso a essa aplicação nas rules

		if !isUserAllowed {
			http.Error(w, "User not allowed to access this application", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
