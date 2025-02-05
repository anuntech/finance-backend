package middlewares

import (
	"net/http"
	"slices"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func IsAllowed(next http.Handler, db *mongo.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		workspaceId := r.Header.Get("workspaceId")
		applicationId := r.Header.Get("applicationId")

		workspaceObjectID, err := primitive.ObjectIDFromHex(workspaceId)
		if err != nil {
			http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
			return
		}

		myApplicationsCollection := db.Collection("myapplications")
		myApplications := myApplicationsCollection.FindOne(helpers.Ctx, bson.M{"workspaceId": workspaceObjectID})
		if myApplications.Err() != nil {
			http.Error(w, "Workspace not found", http.StatusNotFound)
			return
		}

		var myApplication models.MyApplication
		if err := myApplications.Decode(&myApplication); err != nil {
			http.Error(w, "Error decoding my application", http.StatusInternalServerError)
			return
		}

		isWorkspaceAllowedToAccessApplication := slices.Contains(myApplication.AllowedApplicationsId, applicationId)
		if !isWorkspaceAllowedToAccessApplication {
			http.Error(w, "Workspace not allowed to access this application", http.StatusUnauthorized)
			return
		}

		collection := db.Collection("workspaces")
		result := collection.FindOne(helpers.Ctx, bson.M{"_id": workspaceObjectID})

		if result.Err() != nil {
			http.Error(w, "Workspace not found", http.StatusNotFound)
			return
		}

		var workspace models.Workspace
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

		applicationObjectID, err := primitive.ObjectIDFromHex(applicationId)
		if err != nil {
			http.Error(w, "Invalid application ID", http.StatusBadRequest)
			return
		}

		// Check if a normal user has access to the application
		for _, value := range workspace.Rules.AllowedMemberApps {
			if value.AppId == applicationObjectID {
				for _, value := range value.Members {
					if value.MemberId == userObjectID {
						isUserAllowed = true
						break
					}
				}
			}
		}

		if !isUserAllowed {
			http.Error(w, "User not allowed to access this application", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
