//nolint:gocognit
package keeper_test

import (
	gocontext "context"
	"fmt"
	"io/fs"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/samber/lo"

	. "github.com/smartystreets/goconvey/convey"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/axone-protocol/axoned/v8/x/logic"
	"github.com/axone-protocol/axoned/v8/x/logic/keeper"
	logictestutil "github.com/axone-protocol/axoned/v8/x/logic/testutil"
	"github.com/axone-protocol/axoned/v8/x/logic/types"
)

func TestGRPCAsk(t *testing.T) {
	emptySolution := types.Result{}
	Convey("Given a test cases", t, func() {
		cases := []struct {
			program            string
			query              string
			limit              int
			maxResultCount     int
			maxSize            int
			predicateBlacklist []string
			maxGas             uint64
			maxVariables       uint64
			predicateCosts     map[string]uint64
			expectedAnswer     *types.Answer
			expectedError      string
		}{
			{
				program: "foo.",
				query:   "foo.",
				expectedAnswer: &types.Answer{
					Results: []types.Result{emptySolution},
				},
			},
			{
				program:        "father(bob, alice).",
				query:          "father(bob, john).",
				expectedAnswer: &types.Answer{},
			},
			{
				program: "father(bob, alice).",
				query:   "father(bob, X).",
				expectedAnswer: &types.Answer{
					Variables: []string{"X"},
					Results: []types.Result{{Substitutions: []types.Substitution{{
						Variable: "X", Expression: "alice",
					}}}},
				},
			},
			{
				program: `father("bob", "alice").`,
				query:   `father("bob", X).`,
				expectedAnswer: &types.Answer{
					Variables: []string{"X"},
					Results: []types.Result{{Substitutions: []types.Substitution{{
						Variable: "X", Expression: "[a,l,i,c,e]",
					}}}},
				},
			},
			{
				program: "father(bob, alice). father(bob, john).",
				query:   "father(bob, X).",
				expectedAnswer: &types.Answer{
					HasMore:   true,
					Variables: []string{"X"},
					Results: []types.Result{{Substitutions: []types.Substitution{{
						Variable: "X", Expression: "alice",
					}}}},
				},
			},
			{
				program:        "father(bob, alice). father(bob, john).",
				query:          "father(bob, X).",
				maxResultCount: 5,
				expectedAnswer: &types.Answer{
					HasMore:   true,
					Variables: []string{"X"},
					Results: []types.Result{{Substitutions: []types.Substitution{{
						Variable: "X", Expression: "alice",
					}}}},
				},
			},
			{
				program:        "father(bob, alice). father(bob, john).",
				query:          "father(bob, X).",
				limit:          2,
				maxResultCount: 5,
				expectedAnswer: &types.Answer{
					Variables: []string{"X"},
					Results: []types.Result{{Substitutions: []types.Substitution{{
						Variable: "X", Expression: "alice",
					}}}, {Substitutions: []types.Substitution{{
						Variable: "X", Expression: "john",
					}}}},
				},
			},
			{
				program:        "father(bob, alice). father(bob, john).",
				query:          "father(bob, X).",
				limit:          2,
				maxResultCount: 1,
				expectedError:  "query: 2 > MaxResultCount: 1: limit exceeded",
			},
			{
				program:       "father(bob, alice). father(bob, john).",
				query:         "father(bob, X).",
				maxSize:       5,
				expectedError: "query: 15 > MaxSize: 5: limit exceeded",
			},
			{
				query: "block_height(X).",
				expectedAnswer: &types.Answer{
					Variables: []string{"X"},
					Results: []types.Result{{Substitutions: []types.Substitution{{
						Variable: "X", Expression: "0",
					}}}},
				},
			},
			{
				query:         "block_height(X).",
				maxGas:        1000,
				expectedError: "out of gas: logic <ReadPerByte> (1018/1000): limit exceeded",
			},
			{
				query:         "bank_balances(X, Y).",
				maxGas:        3000,
				expectedError: "out of gas: logic <panic: {ValuePerByte}> (3102/3000): limit exceeded",
			},
			{
				query:  "block_height(X).",
				maxGas: 3000,
				predicateCosts: map[string]uint64{
					"block_height/1": 10000,
				},
				expectedError: "out of gas: logic <block_height/1> (11176/3000): limit exceeded",
			},
			{
				query:         "length(List, 100000).",
				maxVariables:  1000,
				expectedError: "maximum number of variables reached: limit exceeded",
			},
			{
				program: "father(bob, 'élodie').",
				query:   "father(bob, X).",
				expectedAnswer: &types.Answer{
					Variables: []string{"X"},
					Results: []types.Result{{Substitutions: []types.Substitution{{
						Variable: "X", Expression: "élodie",
					}}}},
				},
			},
			{
				program: "foo(foo(bar)).",
				query:   "foo(X).",
				expectedAnswer: &types.Answer{
					Variables: []string{"X"},
					Results: []types.Result{{Substitutions: []types.Substitution{{
						Variable: "X", Expression: "foo(bar)",
					}}}},
				},
			},
			{
				program: "father(bob, alice).",
				query:   "father(A, B).",
				expectedAnswer: &types.Answer{
					Variables: []string{"A", "B"},
					Results: []types.Result{{Substitutions: []types.Substitution{{
						Variable: "A", Expression: "bob",
					}, {
						Variable: "B", Expression: "alice",
					}}}},
				},
			},
			{
				program: "father(bob, alice).",
				query:   "father(B, A).",
				expectedAnswer: &types.Answer{
					Variables: []string{"B", "A"},
					Results: []types.Result{{Substitutions: []types.Substitution{{
						Variable: "B", Expression: "bob",
					}, {
						Variable: "A", Expression: "alice",
					}}}},
				},
			},
			{
				program: "father(bob, X) :- true.",
				query:   "father(B, X).",
				expectedAnswer: &types.Answer{
					Variables: []string{"B", "X"},
					Results: []types.Result{{Substitutions: []types.Substitution{{
						Variable: "B", Expression: "bob",
					}}}},
				},
			},
			{
				program: "father(bob, alice).",
				query:   "father(bob, X, O).",
				expectedAnswer: &types.Answer{
					Variables: []string{"X", "O"},
					Results:   []types.Result{{Error: "error(existence_error(procedure,father/3),root)"}},
				},
			},
			{
				program:            "",
				query:              "block_height(X).",
				predicateBlacklist: []string{"block_height/1"},
				expectedAnswer: &types.Answer{
					HasMore:   false,
					Variables: []string{"X"},
					Results:   []types.Result{{Error: "error(permission_error(execute,forbidden_predicate,block_height/1),block_height/1)"}},
				},
			},
			{
				program:       "father°(bob, alice).",
				query:         "father(bob, X).",
				expectedError: "error compiling query: unexpected token: invalid(°): invalid argument",
			},
			{
				program:       "father(bob, alice).",
				query:         "father°(bob, X).",
				expectedError: "error executing query: unexpected token: invalid(°): invalid argument",
			},
			{
				program:       `father("bob", "alice").`,
				query:         `father("bob"", X).`,
				expectedError: "error executing query: EOF: invalid argument",
			},
			{
				program: `
foo(a1).
foo(a2).
foo(a3) :- throw(error(resource_error(foo))).
foo(a4).
`,
				query:          `foo(X).`,
				maxResultCount: 1,
				expectedAnswer: &types.Answer{
					HasMore:   true,
					Variables: []string{"X"},
					Results: []types.Result{
						{Substitutions: []types.Substitution{{Variable: "X", Expression: "a1"}}},
					},
				},
			},
			{
				program: `
foo(a1).
foo(a2).
foo(a3) :- throw(error(resource_error(foo))).
foo(a4).
`,
				query:          `foo(X).`,
				limit:          2,
				maxResultCount: 3,
				expectedAnswer: &types.Answer{
					HasMore:   true,
					Variables: []string{"X"},
					Results: []types.Result{
						{Substitutions: []types.Substitution{{Variable: "X", Expression: "a1"}}},
						{Substitutions: []types.Substitution{{Variable: "X", Expression: "a2"}}},
					},
				},
			},
			{
				program: `
foo(a1).
foo(a2).
foo(a3) :- throw(error(resource_error(foo))).
foo(a4).
`,
				query:          `foo(X).`,
				limit:          3,
				maxResultCount: 5,
				expectedAnswer: &types.Answer{
					Variables: []string{"X"},
					Results: []types.Result{
						{Substitutions: []types.Substitution{{Variable: "X", Expression: "a1"}}},
						{Substitutions: []types.Substitution{{Variable: "X", Expression: "a2"}}},
						{Error: "error(resource_error(foo))"},
					},
				},
			},
			{
				program: `
foo(a1).
foo(a2).
foo(a3) :- throw(error(resource_error(foo))).
foo(a4).
`,
				query:          `foo(X).`,
				limit:          5,
				maxResultCount: 5,
				expectedAnswer: &types.Answer{
					Variables: []string{"X"},
					Results: []types.Result{
						{Substitutions: []types.Substitution{{Variable: "X", Expression: "a1"}}},
						{Substitutions: []types.Substitution{{Variable: "X", Expression: "a2"}}},
						{Error: "error(resource_error(foo))"},
					},
				},
			},
		}

		for nc, tc := range cases {
			Convey(
				fmt.Sprintf("Given test case #%d with program: %v and query: %v", nc, tc.program, tc.query),
				func() {
					encCfg := moduletestutil.MakeTestEncodingConfig(logic.AppModuleBasic{})
					key := storetypes.NewKVStoreKey(types.StoreKey)
					testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))

					// gomock initializations
					ctrl := gomock.NewController(t)
					accountKeeper := logictestutil.NewMockAccountKeeper(ctrl)
					bankKeeper := logictestutil.NewMockBankKeeper(ctrl)
					fsProvider := logictestutil.NewMockFS(ctrl)

					bankKeeper.EXPECT().GetAccountsBalances(gomock.Any()).Do(func(ctx gocontext.Context) []bankypes.Balance {
						sdk.UnwrapSDKContext(ctx).GasMeter().ConsumeGas(2000, "ValuePerByte")
						return nil
					}).AnyTimes()

					logicKeeper := keeper.NewKeeper(
						encCfg.Codec,
						key,
						key,
						authtypes.NewModuleAddress(govtypes.ModuleName),
						accountKeeper,
						bankKeeper,
						func(_ gocontext.Context) fs.FS {
							return fsProvider
						},
					)
					maxResultCount := sdkmath.NewUint(uint64(lo.If(tc.maxResultCount == 0, 1).Else(tc.maxResultCount)))
					maxSize := sdkmath.NewUint(uint64(lo.If(tc.maxSize == 0, 5000).Else(tc.maxSize)))
					params := types.DefaultParams()
					params.Limits.MaxResultCount = &maxResultCount
					params.Limits.MaxSize = &maxSize
					maxVariables := sdkmath.NewUint(tc.maxVariables)
					params.Limits.MaxVariables = &maxVariables
					if tc.predicateBlacklist != nil {
						params.Interpreter.PredicatesFilter.Blacklist = tc.predicateBlacklist
					}
					if tc.predicateCosts != nil {
						predicateCosts := make([]types.PredicateCost, 0, len(tc.predicateCosts))
						for predicate, cost := range tc.predicateCosts {
							cost := sdkmath.NewUint(cost)
							predicateCosts = append(predicateCosts, types.PredicateCost{
								Predicate: predicate,
								Cost:      &cost,
							})
						}
						params.GasPolicy.PredicateCosts = predicateCosts
					}
					err := logicKeeper.SetParams(testCtx.Ctx, params)

					So(err, ShouldBeNil)

					if tc.maxGas != 0 {
						testCtx.Ctx = testCtx.Ctx.WithGasMeter(storetypes.NewGasMeter(tc.maxGas))
					} else {
						testCtx.Ctx = testCtx.Ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())
					}

					Convey("and given a query with program and query to grpc", func() {
						queryHelper := baseapp.NewQueryServerTestHelper(testCtx.Ctx, encCfg.InterfaceRegistry)
						types.RegisterQueryServiceServer(queryHelper, logicKeeper)

						queryClient := types.NewQueryServiceClient(queryHelper)

						var limit *sdkmath.Uint
						if tc.limit != 0 {
							v := sdkmath.NewUint(uint64(tc.limit))
							limit = &v
						}
						query := types.QueryServiceAskRequest{
							Program: tc.program,
							Query:   tc.query,
							Limit:   limit,
						}

						Convey("when the grpc query ask is called", func() {
							result, err := queryClient.Ask(gocontext.Background(), &query)

							Convey("Then it should return the expected answer", func() {
								if tc.expectedError != "" {
									So(err, ShouldNotBeNil)
									So(err.Error(), ShouldEqual, tc.expectedError)
									So(result, ShouldBeNil)
								} else {
									So(err, ShouldBeNil)
									So(result, ShouldNotBeNil)
									So(result.Answer, ShouldResemble, tc.expectedAnswer)
								}
							})
						})
					})
				})
		}
	})
}
