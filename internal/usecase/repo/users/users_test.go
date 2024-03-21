package users

import (
	"context"
	"database/sql"
	repPkg "finance-tg-bot/pkg/repository"
	repoMock "finance-tg-bot/pkg/repository/mocks"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestUser_GetStatus(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repoMock.NewMockUserProvider(ctrl)
	ucaseRepo := New(repo)

	type fields struct {
		repo repPkg.UserProvider
	}
	type args struct {
		ctx      context.Context
		username string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		mockResult *repPkg.DBClient
		mockError  error
		wantStatus bool
		wantErr    bool
	}{
		{
			name:       "active user",
			fields:     fields{repo: ucaseRepo.repo},
			args:       args{ctx: ctx, username: "vasya"},
			mockResult: &repPkg.DBClient{IsActive: sql.NullBool{Bool: true, Valid: true}},
			mockError:  nil,
			wantStatus: true,
			wantErr:    false,
		},
		{
			name:       "inactive user",
			fields:     fields{repo: ucaseRepo.repo},
			args:       args{ctx: ctx, username: "vasya_old"},
			mockResult: &repPkg.DBClient{IsActive: sql.NullBool{Bool: false, Valid: true}},
			mockError:  nil,
			wantStatus: false,
			wantErr:    false,
		},
		{
			name:       "no user",
			fields:     fields{repo: ucaseRepo.repo},
			args:       args{ctx: ctx, username: "petya"},
			mockResult: &repPkg.DBClient{},
			mockError:  sql.ErrNoRows,
			wantStatus: false,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo.EXPECT().GetUserInfo(tt.args.ctx, tt.args.username).Return(tt.mockResult, tt.mockError).Times(1)
			u := &User{
				repo: tt.fields.repo,
			}
			gotStatus, err := u.GetStatus(tt.args.ctx, tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("User.GetStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotStatus != tt.wantStatus {
				t.Errorf("User.GetStatus() = %v, want %v", gotStatus, tt.wantStatus)
			}
		})
	}
}
