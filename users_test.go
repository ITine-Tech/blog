package store

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"testing"
)

func TestUsersPostgresStore_Activate(t *testing.T) {
	type fields struct {
		db *sql.DB
	}
	type args struct {
		ctx   context.Context
		token string
	}

	ctx := context.Background()
	testToken, _ := uuid.NewUUID()
	stringToken := testToken.String()

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "success", args: args{ctx, stringToken}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &UsersPostgresStore{
				db: tt.fields.db,
			}
			if err := s.Activate(tt.args.ctx, tt.args.token); (err != nil) != tt.wantErr {
				t.Errorf("Activate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
