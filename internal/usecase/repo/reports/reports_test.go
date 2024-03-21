package reports

import (
	"context"
	"finance-tg-bot/internal/entity"
	repoMock "finance-tg-bot/pkg/repository/mocks"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestReports_GetStatementTotals(t *testing.T) {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repoMock.NewMockReporter(ctrl)
	ucaseRepo := New(repo)

	type args struct {
		ctx context.Context
		log *slog.Logger
		p   map[string]string
	}
	tests := []struct {
		name       string
		args       args
		mockResult []entity.ReportResult
		mockError  error
		wantRes    string
		wantErr    bool
	}{
		{
			"demo test",
			args{ctx, log, map[string]string{}},
			[]entity.ReportResult{{Name: "boss", Val: 300}, {Name: "vasya", Val: 15}},
			nil,
			`------+----
boss  | 300
vasya |  15
------+----`,
			false,
		},
		{
			"big money",
			args{ctx, log, map[string]string{}},
			[]entity.ReportResult{{Name: "boss", Val: 3000000}, {Name: "vasya", Val: 300}},
			nil,
			`------+----------
boss  | 3 000 000
vasya |       300
------+----------`,
			false,
		},
		{
			"russian report",
			args{ctx, log, map[string]string{}},
			[]entity.ReportResult{{Name: "медведИ", Val: 300000}, {Name: "Наши слоны!", Val: 300}},
			nil,
			`------------+--------
медведИ     | 300 000
Наши слоны! |     300
------------+--------`,
			false,
		},
		{
			"no data",
			args{ctx, log, map[string]string{}},
			nil,
			nil,
			"NO DATA",
			false,
		},
		{
			"demo error",
			args{ctx, log, map[string]string{}},
			nil,
			fmt.Errorf("mockError"),
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo.EXPECT().GetStatementCatTotals(tt.args.ctx, tt.args.p).Return(tt.mockResult, tt.mockError).Times(1)

			gotRes, err := ucaseRepo.GetStatementTotals(tt.args.ctx, tt.args.log, tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("Reports.GetStatementTotals() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println(gotRes)
			if gotRes != tt.wantRes {
				t.Errorf("Reports.GetStatementTotals() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}
