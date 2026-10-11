package api

// addAuthEndpoints adds authentication endpoints to the spec
func (h *OpenAPIHandler) addAuthEndpoints(spec *OpenAPISpec) {
	// User schema — mirrors auth.User (user.go); role/metadata fields are
	// omitempty on the wire
	spec.Components.Schemas["User"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id":             map[string]string{"type": "string", "format": "uuid"},
			"email":          map[string]string{"type": "string", "format": "email"},
			"email_verified": map[string]string{"type": "boolean"},
			"role":           map[string]string{"type": "string", "description": "Assigned when the user has a non-default role"},
			"tenant_id":      map[string]string{"type": "string", "description": "Present when the user belongs to a tenant"},
			"user_metadata":  map[string]interface{}{"type": "object", "additionalProperties": map[string]interface{}{}, "description": "User-editable metadata"},
			"app_metadata":   map[string]interface{}{"type": "object", "additionalProperties": map[string]interface{}{}, "description": "Application/admin-only metadata"},
			"created_at":     map[string]string{"type": "string", "format": "date-time"},
			"updated_at":     map[string]string{"type": "string", "format": "date-time"},
		},
		"required": []string{"id", "email", "email_verified", "created_at", "updated_at"},
	}

	// Token response schema — mirrors auth.SignInResponse
	spec.Components.Schemas["TokenResponse"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"access_token":  map[string]string{"type": "string"},
			"refresh_token": map[string]string{"type": "string"},
			"expires_in":    map[string]string{"type": "integer"},
			"user":          map[string]string{"$ref": "#/components/schemas/User"},
		},
		"required": []string{"access_token", "refresh_token", "expires_in", "user"},
	}

	// Signup response schema — mirrors the auth handler's SignUp response:
	// 201 with tokens, or 201 with requires_email_verification and no tokens
	spec.Components.Schemas["SignUpResponse"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"user":                        map[string]string{"$ref": "#/components/schemas/User"},
			"access_token":                map[string]string{"type": "string", "description": "Absent when email verification is required"},
			"refresh_token":               map[string]string{"type": "string", "description": "Absent when email verification is required"},
			"expires_in":                  map[string]string{"type": "integer", "description": "Access token lifetime in seconds"},
			"requires_email_verification": map[string]string{"type": "boolean", "description": "True when the user must verify their email before signing in"},
			"trust_token":                 map[string]string{"type": "string", "description": "Issued after CAPTCHA verification for subsequent trust checks"},
			"message":                     map[string]string{"type": "string"},
		},
		"required": []string{"user"},
	}

	// Signup request schema — mirrors auth.SignUpRequest
	spec.Components.Schemas["SignupRequest"] = map[string]interface{}{
		"type":     "object",
		"required": []string{"email", "password"},
		"properties": map[string]interface{}{
			"email":              map[string]string{"type": "string", "format": "email"},
			"password":           map[string]string{"type": "string", "minLength": "8"},
			"user_metadata":      map[string]interface{}{"type": "object", "additionalProperties": map[string]interface{}{}},
			"captcha_token":      map[string]string{"type": "string"},
			"challenge_id":       map[string]string{"type": "string", "description": "Challenge ID from the pre-flight CAPTCHA check"},
			"device_fingerprint": map[string]string{"type": "string"},
		},
	}

	// Signin request schema — mirrors auth.SignInRequest
	spec.Components.Schemas["SigninRequest"] = map[string]interface{}{
		"type":     "object",
		"required": []string{"email", "password"},
		"properties": map[string]interface{}{
			"email":              map[string]string{"type": "string", "format": "email"},
			"password":           map[string]string{"type": "string"},
			"captcha_token":      map[string]string{"type": "string"},
			"challenge_id":       map[string]string{"type": "string", "description": "Challenge ID from the pre-flight CAPTCHA check"},
			"device_fingerprint": map[string]string{"type": "string"},
		},
	}

	// Refresh request schema
	spec.Components.Schemas["RefreshRequest"] = map[string]interface{}{
		"type":     "object",
		"required": []string{"refresh_token"},
		"properties": map[string]interface{}{
			"refresh_token": map[string]string{"type": "string"},
		},
	}

	// Delete account request schema (self-service account deletion)
	spec.Components.Schemas["DeleteAccountRequest"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"password": map[string]string{
				"type":        "string",
				"description": "Required when the account has a password credential; omit for OAuth-only accounts",
			},
		},
	}

	// Magic link request schema
	spec.Components.Schemas["MagicLinkRequest"] = map[string]interface{}{
		"type":     "object",
		"required": []string{"email"},
		"properties": map[string]interface{}{
			"email":         map[string]string{"type": "string", "format": "email"},
			"captcha_token": map[string]string{"type": "string"},
		},
	}

	// Error response schema
	spec.Components.Schemas["Error"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"error": map[string]string{"type": "string"},
		},
	}

	// POST /api/v1/auth/signup
	spec.Paths["/api/v1/auth/signup"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Sign up a new user",
			Description: "Create a new user account with email and password",
			OperationID: "auth_signup",
			Tags:        []string{"Authentication"},
			RequestBody: &OpenAPIRequestBody{
				Required: true,
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]string{"$ref": "#/components/schemas/SignupRequest"},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"201": {
					Description: "User created. Returns tokens when email verification is disabled, otherwise requires_email_verification=true with no tokens.",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/SignUpResponse"},
						},
					},
				},
				"400": {
					Description: "Invalid request",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Error"},
						},
					},
				},
				"403": {
					Description: "Signup disabled or CAPTCHA/challenge verification failed",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Error"},
						},
					},
				},
			},
		},
	}

	// POST /api/v1/auth/signin
	spec.Paths["/api/v1/auth/signin"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Sign in a user",
			Description: "Authenticate with email and password to get access tokens",
			OperationID: "auth_signin",
			Tags:        []string{"Authentication"},
			RequestBody: &OpenAPIRequestBody{
				Required: true,
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]string{"$ref": "#/components/schemas/SigninRequest"},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Successfully authenticated",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/TokenResponse"},
						},
					},
				},
				"401": {
					Description: "Invalid credentials",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Error"},
						},
					},
				},
			},
		},
	}

	// POST /api/v1/auth/signout
	spec.Paths["/api/v1/auth/signout"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Sign out a user",
			Description: "Invalidate the current session",
			OperationID: "auth_signout",
			Tags:        []string{"Authentication"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Successfully signed out",
				},
				"401": {
					Description: "Unauthorized",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Error"},
						},
					},
				},
			},
		},
	}

	// DELETE /api/v1/auth/account
	spec.Paths["/api/v1/auth/account"] = OpenAPIPath{
		"delete": OpenAPIOperation{
			Summary:     "Delete own account",
			Description: "Permanently delete the authenticated user's account with all sessions and tokens, owned knowledge-base documents and storage object metadata. Requires the account password when the user has one; OAuth-only users may omit the body.",
			OperationID: "auth_delete_account",
			Tags:        []string{"Authentication"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			RequestBody: &OpenAPIRequestBody{
				Required: false,
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]string{"$ref": "#/components/schemas/DeleteAccountRequest"},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"204": {
					Description: "Account deleted",
				},
				"401": {
					Description: "Unauthorized",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Error"},
						},
					},
				},
				"403": {
					Description: "Password missing or incorrect",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Error"},
						},
					},
				},
				"404": {
					Description: "User not found",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Error"},
						},
					},
				},
			},
		},
	}

	// POST /api/v1/auth/refresh
	spec.Paths["/api/v1/auth/refresh"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Refresh access token",
			Description: "Get a new access token using a refresh token",
			OperationID: "auth_refresh",
			Tags:        []string{"Authentication"},
			RequestBody: &OpenAPIRequestBody{
				Required: true,
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]string{"$ref": "#/components/schemas/RefreshRequest"},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "New access token issued",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/TokenResponse"},
						},
					},
				},
				"401": {
					Description: "Invalid refresh token",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Error"},
						},
					},
				},
			},
		},
	}

	// GET /api/v1/auth/user
	spec.Paths["/api/v1/auth/user"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "Get current user",
			Description: "Get the authenticated user's information",
			OperationID: "auth_get_user",
			Tags:        []string{"Authentication"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "User information",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/User"},
						},
					},
				},
				"401": {
					Description: "Unauthorized",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Error"},
						},
					},
				},
			},
		},
		"patch": OpenAPIOperation{
			Summary:     "Update current user",
			Description: "Update the authenticated user's information",
			OperationID: "auth_update_user",
			Tags:        []string{"Authentication"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			RequestBody: &OpenAPIRequestBody{
				Required: true,
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]string{"$ref": "#/components/schemas/User"},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "User updated successfully",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/User"},
						},
					},
				},
				"401": {
					Description: "Unauthorized",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Error"},
						},
					},
				},
			},
		},
	}

	// POST /api/v1/auth/magiclink
	spec.Paths["/api/v1/auth/magiclink"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Request magic link",
			Description: "Send a magic link to the user's email for passwordless authentication",
			OperationID: "auth_magiclink",
			Tags:        []string{"Authentication"},
			RequestBody: &OpenAPIRequestBody{
				Required: true,
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]string{"$ref": "#/components/schemas/MagicLinkRequest"},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Magic link sent successfully",
				},
				"400": {
					Description: "Invalid request",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Error"},
						},
					},
				},
			},
		},
	}

	// GET /api/v1/auth/magiclink/verify
	spec.Paths["/api/v1/auth/magiclink/verify"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "Verify magic link",
			Description: "Verify a magic link token and authenticate the user",
			OperationID: "auth_magiclink_verify",
			Tags:        []string{"Authentication"},
			Parameters: []OpenAPIParameter{
				{
					Name:        "token",
					In:          "query",
					Description: "Magic link token",
					Required:    true,
					Schema:      map[string]string{"type": "string"},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Successfully authenticated",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/TokenResponse"},
						},
					},
				},
				"401": {
					Description: "Invalid or expired token",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Error"},
						},
					},
				},
			},
		},
	}
}
