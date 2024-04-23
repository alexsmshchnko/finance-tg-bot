package repo

import (
	"context"
	"database/sql"
	"finance-tg-bot/internal/entity"
	repPkg "finance-tg-bot/pkg/repository"
	repoMock "finance-tg-bot/pkg/repository/mocks"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
)

func TestRepo_PostDocument(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rpMck := repoMock.NewMockDocProcessor(ctrl)

	type fields struct {
		repo repPkg.DocProcessor
	}
	type args struct {
		ctx context.Context
		doc *entity.Document
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		mockDoc    *repPkg.DBDocument
		mockResult error
		wantErr    bool
	}{
		{
			name:   "demo posting",
			fields: fields{repo: rpMck},
			args: args{ctx: context.Background(), doc: &entity.Document{
				RecTime:     time.Unix(int64(1405544146), 0),
				Category:    "test",
				Amount:      int64(24234),
				Description: "testim",
				MsgID:       "1600",
				ChatID:      "1234",
				UserId:      1,
				Direction:   int8(-1),
			}},
			mockDoc: &repPkg.DBDocument{
				RecDate:     sql.NullTime{Time: time.Unix(int64(1405544146), 0), Valid: true},
				TransDate:   sql.NullTime{Time: time.Time{}, Valid: false},
				Category:    sql.NullString{String: "test", Valid: true},
				Amount:      sql.NullInt64{Int64: 24234, Valid: true},
				Description: sql.NullString{String: "testim", Valid: true},
				MsgID:       sql.NullString{String: "1600", Valid: true},
				ChatID:      sql.NullString{String: "1234", Valid: true},
				UserId:      sql.NullInt64{Int64: 1, Valid: true},
				Direction:   sql.NullInt16{Int16: 0, Valid: false},
			},
			mockResult: nil,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rpMck.EXPECT().PostDocument(tt.args.ctx, tt.mockDoc).Return(nil).Times(1)
			rpMck.EXPECT().GetDocumentCategories(tt.args.ctx, tt.args.doc.UserId, "").Times(1)
			rpMck.EXPECT().GetDocumentSubCategories(tt.args.ctx, tt.args.doc.UserId, tt.args.doc.Category).Times(1)

			if err := New(tt.fields.repo).PostDocument(tt.args.ctx, tt.args.doc); (err != nil) != tt.wantErr {
				t.Errorf("Repo.PostDocument() error = %v, wantErr %v", err, tt.wantErr)
			}
			time.Sleep(2 * time.Second) //wait goroutine
		})
	}
}

// func TestRepo_DeleteDocument(t *testing.T) {
// 	type fields struct {
// 		repo  repPkg.DocProcessor
// 		cache map[string]categories
// 	}
// 	type args struct {
// 		ctx context.Context
// 		doc *entity.Document
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r := &Repo{
// 				repo:  tt.fields.repo,
// 				cache: tt.fields.cache,
// 			}
// 			if err := r.DeleteDocument(tt.args.ctx, tt.args.doc); (err != nil) != tt.wantErr {
// 				t.Errorf("Repo.DeleteDocument() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

func TestRepo_GetCategories(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rpMck := repoMock.NewMockDocProcessor(ctrl)

	type fields struct {
		repo repPkg.DocProcessor
		// cache map[string]categories
	}
	type args struct {
		ctx     context.Context
		user_id int
		limit   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantCat []entity.TransCatLimit
		wantErr bool
	}{
		{
			name:   "3 cats",
			fields: fields{repo: rpMck},
			args:   args{ctx: context.Background(), user_id: 1, limit: "balance"},
			wantCat: []entity.TransCatLimit{
				{Category: sql.NullString{String: "apple", Valid: true},
					Direction: sql.NullInt16{Int16: -1, Valid: true},
					UserId:    sql.NullInt64{Int64: 1, Valid: true},
					Active:    sql.NullBool{Bool: true, Valid: true},
					Limit:     sql.NullInt64{Int64: 100, Valid: true},
					Balance:   sql.NullInt64{Int64: 20, Valid: true},
				},
				{Category: sql.NullString{String: "banana", Valid: true},
					Direction: sql.NullInt16{Int16: 1, Valid: true},
					UserId:    sql.NullInt64{Int64: 1, Valid: true},
					Active:    sql.NullBool{Bool: true, Valid: true},
					Limit:     sql.NullInt64{Int64: 0, Valid: false},
					Balance:   sql.NullInt64{Int64: 250, Valid: true},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rpMck.EXPECT().GetDocumentCategories(tt.args.ctx, tt.args.user_id, tt.args.limit).Return(tt.wantCat, nil).Times(1)
			r := New(tt.fields.repo)

			gotCat, err := r.GetCategories(tt.args.ctx, tt.args.user_id, tt.args.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repo.GetCategories() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotCat, tt.wantCat) {
				t.Errorf("Repo.GetCategories() = %v, want %v", gotCat, tt.wantCat)
			}
		})
	}
}

func TestRepo_GetSubCategories(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rpMck := repoMock.NewMockDocProcessor(ctrl)

	type fields struct {
		repo repPkg.DocProcessor
		// cache map[string]categories
	}
	type args struct {
		ctx       context.Context
		user_id   int
		trans_cat string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		repoCat []entity.TransCatLimit
		wantCat []string
		wantErr bool
	}{
		{
			name:   "3 cats",
			fields: fields{repo: rpMck},
			args:   args{ctx: context.Background(), user_id: 1, trans_cat: "food"},
			repoCat: []entity.TransCatLimit{
				{
					Category:  sql.NullString{String: "food", Valid: true},
					Direction: sql.NullInt16{Int16: -1, Valid: true},
					UserId:    sql.NullInt64{Int64: 1, Valid: true},
					Active:    sql.NullBool{Bool: true, Valid: true},
					Limit:     sql.NullInt64{Int64: 100, Valid: true},
					Balance:   sql.NullInt64{Int64: 20, Valid: true},
				},
			},
			wantCat: []string{"apple", "banana", "orange"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rpMck.EXPECT().GetDocumentSubCategories(tt.args.ctx, tt.args.user_id, tt.args.trans_cat).Return(tt.wantCat, nil).Times(1)

			r := New(tt.fields.repo)
			gotCat, err := r.GetSubCategories(tt.args.ctx, tt.args.user_id, tt.args.trans_cat)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repo.GetSubCategories() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotCat, tt.wantCat) {
				t.Errorf("Repo.GetSubCategories() = %v, wa nt %v", gotCat, tt.wantCat)
			}
		})
	}
}

// func TestRepo_EditCategory(t *testing.T) {
// 	type fields struct {
// 		repo  repPkg.DocProcessor
// 		cache map[string]categories
// 	}
// 	type args struct {
// 		ctx    context.Context
// 		tc     entity.TransCatLimit
// 		client string
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r := &Repo{
// 				repo:  tt.fields.repo,
// 				cache: tt.fields.cache,
// 			}
// 			if err := r.EditCategory(tt.args.ctx, tt.args.tc, tt.args.client); (err != nil) != tt.wantErr {
// 				t.Errorf("Repo.EditCategory() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }
