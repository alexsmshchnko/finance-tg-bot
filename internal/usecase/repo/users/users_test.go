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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rpMck := repoMock.NewMockUserProvider(ctrl)
	ucaseRepo := New(rpMck)

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
		wantId     int
		wantStatus bool
		wantErr    bool
	}{
		{
			name:   "active user",
			fields: fields{repo: ucaseRepo.repo},
			args:   args{ctx: context.Background(), username: "vasya"},
			mockResult: &repPkg.DBClient{
				ID:       sql.NullInt64{Int64: 1, Valid: true},
				IsActive: sql.NullBool{Bool: true, Valid: true},
			},
			mockError:  nil,
			wantId:     1,
			wantStatus: true,
			wantErr:    false,
		},
		{
			name:   "inactive user",
			fields: fields{repo: ucaseRepo.repo},
			args:   args{ctx: context.Background(), username: "vasya_old"},
			mockResult: &repPkg.DBClient{
				ID:       sql.NullInt64{Int64: 1, Valid: true},
				IsActive: sql.NullBool{Bool: false, Valid: true},
			},
			mockError:  nil,
			wantId:     1,
			wantStatus: false,
			wantErr:    false,
		},
		{
			name:       "no user",
			fields:     fields{repo: ucaseRepo.repo},
			args:       args{ctx: context.Background(), username: "petya"},
			mockResult: &repPkg.DBClient{},
			mockError:  sql.ErrNoRows,
			wantId:     0,
			wantStatus: false,
			wantErr:    false,
		},
		{
			name:       "some error",
			fields:     fields{repo: ucaseRepo.repo},
			args:       args{ctx: context.Background(), username: "petya"},
			mockResult: nil,
			mockError:  sql.ErrTxDone,
			wantId:     0,
			wantStatus: false,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rpMck.EXPECT().GetUserInfo(tt.args.ctx, tt.args.username).Return(tt.mockResult, tt.mockError).Times(1)

			gotId, gotStatus, err := New(tt.fields.repo).GetStatus(tt.args.ctx, tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("User.GetStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotId != tt.wantId {
				t.Errorf("User.GetStatus() = %v, want %v", gotId, tt.wantId)
			}
			if gotStatus != tt.wantStatus {
				t.Errorf("User.GetStatus() = %v, want %v", gotStatus, tt.wantStatus)
			}
		})
	}
}
