package repo

import (
	"context"
	"errors"
	"finance-tg-bot/internal/entity"
	repPkg "finance-tg-bot/pkg/repository"
	repoMock "finance-tg-bot/pkg/repository/mocks"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
)

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rpMck := repoMock.NewMockDocProcessor(ctrl)

	type args struct {
		repPkg repPkg.DocProcessor
	}
	tests := []struct {
		name string
		args args
		want *Repo
	}{
		{
			name: "new check",
			args: args{repPkg: rpMck},
			want: &Repo{repo: rpMck, cacheCats: map[int][]entity.TransCatLimit{}, cacheSubCats: map[int]map[string][]string{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.repPkg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepo_clearCache(t *testing.T) {
	type fields struct {
		// repo         repPkg.DocProcessor
		cacheCats    map[int][]entity.TransCatLimit
		cacheSubCats map[int]map[string][]string
	}
	type args struct {
		user_id   int
		tranc_cat string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "only cats",
			fields: fields{cacheCats: map[int][]entity.TransCatLimit{}, cacheSubCats: map[int]map[string][]string{}},
			args:   args{user_id: 1, tranc_cat: ""},
		},
		{
			name:   "plus subcats",
			fields: fields{cacheCats: map[int][]entity.TransCatLimit{}, cacheSubCats: map[int]map[string][]string{}},
			args:   args{user_id: 1, tranc_cat: "food"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Repo{cacheCats: tt.fields.cacheCats, cacheSubCats: tt.fields.cacheSubCats}
			r.clearCache(tt.args.user_id, tt.args.tranc_cat)
		})
	}
}

func TestRepo_PostDocument(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rpMck := repoMock.NewMockDocProcessor(ctrl)
	ctx := context.Background()
	tm := time.Now()

	type fields struct {
		repo         repPkg.DocProcessor
		cacheCats    map[int][]entity.TransCatLimit
		cacheSubCats map[int]map[string][]string
	}
	type args struct {
		ctx context.Context
		doc *entity.Document
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		mockDoc *entity.Document
		wantErr bool
	}{
		{
			name: "demo posting",
			fields: fields{repo: rpMck,
				cacheCats: map[int][]entity.TransCatLimit{
					1: {{Category: "debit", Direction: -1, UserId: 1},
						{Category: "credit", Direction: 1, UserId: 1}}},
				cacheSubCats: make(map[int]map[string][]string)},
			args: args{ctx: ctx,
				doc: &entity.Document{
					RecTime:     time.Unix(int64(1405544146), 0),
					Category:    "debit",
					Amount:      int64(24234),
					Description: "testim",
					MsgID:       "1600",
					ChatID:      "1234",
					UserId:      1}},
			mockDoc: &entity.Document{
				RecTime:     time.Unix(int64(1405544146), 0),
				TransDate:   tm,
				Category:    "debit",
				Amount:      24234,
				Description: "testim",
				MsgID:       "1600",
				ChatID:      "1234",
				UserId:      1,
				Direction:   -1,
			},
			wantErr: false,
		},
		{
			name: "posting with no init cache",
			fields: fields{repo: rpMck,
				cacheCats:    make(map[int][]entity.TransCatLimit),
				cacheSubCats: make(map[int]map[string][]string)},
			args: args{ctx: ctx,
				doc: &entity.Document{
					RecTime:     time.Unix(int64(1405544156), 0),
					Category:    "deposit",
					Amount:      int64(10000),
					Description: "testim deposit",
					MsgID:       "1610",
					ChatID:      "1234",
					UserId:      1}},
			mockDoc: &entity.Document{
				RecTime:     time.Unix(int64(1405544156), 0),
				TransDate:   tm,
				Category:    "deposit",
				Amount:      10000,
				Description: "testim deposit",
				MsgID:       "1610",
				ChatID:      "1234",
				UserId:      1,
				Direction:   0,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.fields.repo)

			switch tt.name {
			case "demo posting":
				r.cacheCats = tt.fields.cacheCats
				rpMck.EXPECT().GetDocumentCategories(tt.args.ctx, tt.args.doc.UserId).Times(1)
			case "posting with no init cache":
				rpMck.EXPECT().GetDocumentCategories(tt.args.ctx, tt.args.doc.UserId).Times(2)
			}
			rpMck.EXPECT().PostDocument(tt.args.ctx, tt.mockDoc).Return(nil).Times(1)
			rpMck.EXPECT().GetDocumentSubCategories(tt.args.ctx, tt.args.doc.UserId, tt.args.doc.Category).Times(1)

			tt.args.doc.TransDate = tm
			if err := r.PostDocument(tt.args.ctx, tt.args.doc); (err != nil) != tt.wantErr {
				t.Errorf("Repo.PostDocument() error = %v, wantErr %v", err, tt.wantErr)
			}
			time.Sleep(1 * time.Second) //wait goroutine
		})
	}
}

func TestRepo_DeleteDocument(t *testing.T) {
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
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "demo1",
			fields: fields{repo: rpMck},
			args: args{ctx: context.Background(), doc: &entity.Document{
				MsgID:  "123",
				ChatID: "321",
				UserId: 1,
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rpMck.EXPECT().DeleteDocument(context.Background(), tt.args.doc).Times(1)
			r := &Repo{repo: tt.fields.repo}
			if err := r.DeleteDocument(tt.args.ctx, tt.args.doc); (err != nil) != tt.wantErr {
				t.Errorf("Repo.DeleteDocument() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRepo_GetCategories(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rpMck := repoMock.NewMockDocProcessor(ctrl)
	ctx := context.Background()

	type fields struct {
		repo      repPkg.DocProcessor
		cacheCats map[int][]entity.TransCatLimit
		// cacheSubCats map[int]map[string][]string
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
			name:   "4 cats with no cache",
			fields: fields{repo: rpMck, cacheCats: make(map[int][]entity.TransCatLimit)},
			args:   args{ctx: ctx, user_id: 1, limit: "balance"},
			wantCat: []entity.TransCatLimit{
				{Category: "apple", Direction: -1, UserId: 1, Active: true, Limit: 100, Balance: 20},
				{Category: "pinapple", Direction: -1, UserId: 1, Active: true, Limit: 150, Balance: 30},
				{Category: "banana", Direction: 1, UserId: 1, Active: true, Limit: 0, Balance: 250},
				{Category: "carrot", Direction: 0, UserId: 1, Active: true, Limit: 0, Balance: 80},
			},
			wantErr: false,
		},
		{
			name: "3 cats cached",
			fields: fields{repo: rpMck, cacheCats: map[int][]entity.TransCatLimit{1: {
				{Category: "mango", Direction: -1, UserId: 1, Active: true, Limit: 300, Balance: 20},
				{Category: "orange", Direction: 1, UserId: 1, Active: true, Limit: 0, Balance: 100},
				{Category: "carrot", Direction: 0, UserId: 1, Active: true, Limit: 40, Balance: 60},
			}}},
			args: args{ctx: ctx, user_id: 1, limit: "balance"},
			wantCat: []entity.TransCatLimit{
				{Category: "mango", Direction: -1, UserId: 1, Active: true, Limit: 300, Balance: 20},
				{Category: "orange", Direction: 1, UserId: 1, Active: true, Limit: 0, Balance: 100},
				{Category: "carrot", Direction: 0, UserId: 1, Active: true, Limit: 40, Balance: 60},
			},
			wantErr: false,
		},
		{
			name:    "some error",
			fields:  fields{repo: rpMck, cacheCats: make(map[int][]entity.TransCatLimit)},
			args:    args{ctx: ctx, user_id: 1, limit: ""},
			wantCat: nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.name {
			case "4 cats with no cache":
				rpMck.EXPECT().GetDocumentCategories(tt.args.ctx, tt.args.user_id).Return(tt.wantCat, nil).Times(1)
			case "2 cats cached":
			case "some error":
				rpMck.EXPECT().GetDocumentCategories(tt.args.ctx, tt.args.user_id).Return(nil, errors.New("test")).Times(1)
			}

			r := &Repo{repo: tt.fields.repo, cacheCats: tt.fields.cacheCats}

			gotCat, err := r.GetCategories(tt.args.ctx, tt.args.user_id)
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
	ctx := context.Background()

	type fields struct {
		repo repPkg.DocProcessor
		// cacheCats map[int][]entity.TransCatLimit
		cacheSubCats map[int]map[string][]string
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
		wantCat []string
		wantErr bool
	}{
		{
			name:    "2 subcats no cache",
			fields:  fields{repo: rpMck, cacheSubCats: make(map[int]map[string][]string)},
			args:    args{ctx: ctx, user_id: 1, trans_cat: "food"},
			wantCat: []string{"apple", "banana"},
			wantErr: false,
		},
		{
			name: "3 subcats cached",
			fields: fields{repo: rpMck,
				cacheSubCats: map[int]map[string][]string{1: {"food": []string{"apple", "banana", "orange"}}}},
			args:    args{ctx: ctx, user_id: 1, trans_cat: "food"},
			wantCat: []string{"apple", "banana", "orange"},
			wantErr: false,
		},
		{
			name:    "some error",
			fields:  fields{repo: rpMck, cacheSubCats: make(map[int]map[string][]string)},
			args:    args{ctx: ctx, user_id: 1, trans_cat: "food"},
			wantCat: nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.name {
			case "2 subcats no cache":
				rpMck.EXPECT().GetDocumentSubCategories(tt.args.ctx, tt.args.user_id, tt.args.trans_cat).Return(tt.wantCat, nil).Times(1)
			case "3 subcats cached":
			case "some error":
				rpMck.EXPECT().GetDocumentSubCategories(tt.args.ctx, tt.args.user_id, tt.args.trans_cat).Return(nil, errors.New("test")).Times(1)
			}

			r := &Repo{repo: rpMck, cacheSubCats: tt.fields.cacheSubCats}

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

func TestRepo_EditCategory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rpMck := repoMock.NewMockDocProcessor(ctrl)
	ctx := context.Background()

	type fields struct {
		repo      repPkg.DocProcessor
		cacheCats map[int][]entity.TransCatLimit
		// cacheSubCats map[int]map[string][]string
	}
	type args struct {
		ctx context.Context
		tc  *entity.TransCatLimit
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "set inactive",
			fields:  fields{repo: rpMck, cacheCats: make(map[int][]entity.TransCatLimit)},
			args:    args{ctx: ctx, tc: &entity.TransCatLimit{Category: "food", Direction: -1, UserId: 1, Active: false}},
			wantErr: false,
		},
		{
			name:    "set limit",
			fields:  fields{repo: rpMck, cacheCats: make(map[int][]entity.TransCatLimit)},
			args:    args{ctx: ctx, tc: &entity.TransCatLimit{Category: "food", Direction: -1, UserId: 1, Active: true, Limit: 100}},
			wantErr: false,
		},
		{
			name:    "some error",
			fields:  fields{repo: rpMck, cacheCats: make(map[int][]entity.TransCatLimit)},
			args:    args{ctx: ctx, tc: &entity.TransCatLimit{Category: "food", Direction: -1, UserId: 1, Active: false}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.name {
			case "set inactive":
				rpMck.EXPECT().EditCategory(tt.args.ctx, tt.args.tc).Return(nil).Times(1)
				rpMck.EXPECT().GetDocumentCategories(tt.args.ctx, tt.args.tc.UserId).Return(nil, nil).Times(1)
			case "set limit":
				rpMck.EXPECT().EditCategory(tt.args.ctx, tt.args.tc).Return(nil).Times(1)
				rpMck.EXPECT().GetDocumentCategories(tt.args.ctx, tt.args.tc.UserId).Return(nil, nil).Times(1)
			case "some error":
				rpMck.EXPECT().EditCategory(tt.args.ctx, tt.args.tc).Return(errors.New("test")).Times(1)
				rpMck.EXPECT().GetDocumentCategories(tt.args.ctx, tt.args.tc.UserId).Return(nil, nil).Times(1)
			}

			r := &Repo{repo: tt.fields.repo, cacheCats: tt.fields.cacheCats}

			if err := r.EditCategory(tt.args.ctx, tt.args.tc); (err != nil) != tt.wantErr {
				t.Errorf("Repo.EditCategory() error = %v, wantErr %v", err, tt.wantErr)
			}
			time.Sleep(2 * time.Second) //wait goroutine
		})
	}
}
