package main

import (
	"os"
	"path"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/lcd"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	app "github.com/Sultanashazath/nameservice"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/cli"
)

func main() {
	cobra.EnableCommandSorting = false

	cdc := app.MakeCodec()

	// Read in the configuration file for the sdk
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(sdk.Bech32PrefixAccAddr, sdk.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(sdk.Bech32PrefixValAddr, sdk.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(sdk.Bech32PrefixConsAddr, sdk.Bech32PrefixConsPub)
	config.Seal()

	rootCmd := &cobra.Command{
		Use:   "nscli",
		Short: "nameservice Client",
	}

	// Add --chain-id to persistent flags and mark it required
	rootCmd.PersistentFlags().String(flags.FlagChainID, "", "Chain ID of tendermint node")
	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		return initConfig(rootCmd)
	}

	// Construct Root Command
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		client.ConfigCmd(app.DefaultCLIHome),
		queryCmd(cdc),
		txCmd(cdc),
		flags.LineBreak,
		lcd.ServeCommand(cdc, registerRoutes),
		flags.LineBreak,
		keys.Commands(),
		flags.LineBreak,
		version.Cmd,
		flags.NewCompletionCmd(rootCmd, true),
	)

	executor := cli.PrepareMainCmd(rootCmd, "NS", app.DefaultCLIHome)
	err := executor.Execute()

	if err != nil {
		panic(err)
	}
}

func registerRoutes(rs *lcd.RestServer) {
	client.RegisterRoutes(rs.CliCtx, rs.Mux)
	authrest.RegisterTxRoutes(rs.CliCtx, rs.Mux)
	app.ModuleBasics.RegisterRESTRoutes(rs.CliCtx, rs.Mux)
}

func queryCmd(cdc *amino.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:     "query",
		Aliases: []string{"q"},
		Short:   "Querying subcommands",
	}

	queryCmd.AddCommand(
		authcmd.GetAccountCmd(cdc),
		flags.LineBreak,
		rpc.ValidatorCommand(cdc),
		rpc.BlockCommand(),
		authcmd.QueryTxsByEventsCmd(cdc),
		authcmd.QueryTxCmd(cdc),
		flags.LineBreak,
	)

	// add modules' query commands
	app.ModuleBasics.AddQueryCommands(queryCmd, cdc)

	return queryCmd
}

func txCmd(cdc *amino.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "tx",
		Short: "Transactions subcommands",
	}

	txCmd.AddCommand(
		bankcmd.SendTxCmd(cdc),
		flags.LineBreak,
		authcmd.GetSignCommand(cdc),
		authcmd.GetMultiSignCommand(cdc),
		flags.LineBreak,
		authcmd.GetBroadcastCommand(cdc),
		authcmd.GetEncodeCommand(cdc),
		authcmd.GetDecodeCommand(cdc),
		flags.LineBreak,
	)

	// add modules' tx commands
	app.ModuleBasics.AddTxCommands(txCmd, cdc)

	// remove auth and bank commands as they're mounted under the root tx command
	var cmdsToRemove []*cobra.Command

	for _, cmd := range txCmd.Commands() {
		if cmd.Use == auth.ModuleName || cmd.Use == bank.ModuleName {
			cmdsToRemove = append(cmdsToRemove, cmd)
		}
	}

	txCmd.RemoveCommand(cmdsToRemove...)

	return txCmd
}

func initConfig(cmd *cobra.Command) error {
	home, err := cmd.PersistentFlags().GetString(cli.HomeFlag)
	if err != nil {
		return err
	}

	cfgFile := path.Join(home, "config", "config.toml")
	if _, err := os.Stat(cfgFile); err == nil {
		viper.SetConfigFile(cfgFile)

		if err := viper.ReadInConfig(); err != nil {
			return err
		}
	}
	if err := viper.BindPFlag(flags.FlagChainID, cmd.PersistentFlags().Lookup(flags.FlagChainID)); err != nil {
		return err
	}
	if err := viper.BindPFlag(cli.EncodingFlag, cmd.PersistentFlags().Lookup(cli.EncodingFlag)); err != nil {
		return err
	}
	return viper.BindPFlag(cli.OutputFlag, cmd.PersistentFlags().Lookup(cli.OutputFlag))
}



// package main

// import (
// 	"encoding/json"
// 	"io"

// 	"github.com/cosmos/cosmos-sdk/server"
// 	"github.com/cosmos/cosmos-sdk/x/staking"

// 	"github.com/spf13/cobra"
// 	"github.com/spf13/viper"
// 	"github.com/tendermint/tendermint/libs/cli"
// 	"github.com/tendermint/tendermint/libs/log"

// 	"github.com/cosmos/cosmos-sdk/baseapp"
// 	"github.com/cosmos/cosmos-sdk/client/debug"
// 	"github.com/cosmos/cosmos-sdk/client/flags"
// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	"github.com/cosmos/cosmos-sdk/x/auth"
// 	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
// 	app "github.com/cosmos/sdk-tutorials/nameservice"
// 	abci "github.com/tendermint/tendermint/abci/types"
// 	tmtypes "github.com/tendermint/tendermint/types"
// 	dbm "github.com/tendermint/tm-db"
// )

// func main() {
// 	cobra.EnableCommandSorting = false

// 	cdc := app.MakeCodec()

// 	config := sdk.GetConfig()
// 	config.SetBech32PrefixForAccount(sdk.Bech32PrefixAccAddr, sdk.Bech32PrefixAccPub)
// 	config.SetBech32PrefixForValidator(sdk.Bech32PrefixValAddr, sdk.Bech32PrefixValPub)
// 	config.SetBech32PrefixForConsensusNode(sdk.Bech32PrefixConsAddr, sdk.Bech32PrefixConsPub)
// 	config.Seal()

// 	ctx := server.NewDefaultContext()

// 	rootCmd := &cobra.Command{
// 		Use:               "nsd",
// 		Short:             "nameservice App Daemon (server)",
// 		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
// 	}
// 	// CLI commands to initialize the chain
// 	rootCmd.AddCommand(genutilcli.InitCmd(ctx, cdc, app.ModuleBasics, app.DefaultNodeHome))
// 	rootCmd.AddCommand(genutilcli.CollectGenTxsCmd(ctx, cdc, auth.GenesisAccountIterator{}, app.DefaultNodeHome))
// 	rootCmd.AddCommand(genutilcli.MigrateGenesisCmd(ctx, cdc))
// 	rootCmd.AddCommand(
// 		genutilcli.GenTxCmd(
// 			ctx, cdc, app.ModuleBasics, staking.AppModuleBasic{},
// 			auth.GenesisAccountIterator{}, app.DefaultNodeHome, app.DefaultCLIHome,
// 		),
// 	)
// 	rootCmd.AddCommand(genutilcli.ValidateGenesisCmd(ctx, cdc, app.ModuleBasics))
// 	// AddGenesisAccountCmd allows users to add accounts to the genesis file
// 	rootCmd.AddCommand(AddGenesisAccountCmd(ctx, cdc, app.DefaultNodeHome, app.DefaultCLIHome))
// 	rootCmd.AddCommand(flags.NewCompletionCmd(rootCmd, true))
// 	rootCmd.AddCommand(debug.Cmd(cdc))

// 	server.AddCommands(ctx, cdc, rootCmd, newApp, exportAppStateAndTMValidators)

// 	// prepare and add flags
// 	executor := cli.PrepareBaseCmd(rootCmd, "NS", app.DefaultNodeHome)
// 	err := executor.Execute()
// 	if err != nil {
// 		panic(err)
// 	}
// }

// func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer) abci.Application {
// 	return app.NewNameServiceApp(logger, db, baseapp.SetMinGasPrices(viper.GetString(server.FlagMinGasPrices)))
// }

// func exportAppStateAndTMValidators(
// 	logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailWhiteList []string,
// ) (json.RawMessage, []tmtypes.GenesisValidator, error) {

// 	if height != -1 {
// 		nsApp := app.NewNameServiceApp(logger, db)
// 		err := nsApp.LoadHeight(height)
// 		if err != nil {
// 			return nil, nil, err
// 		}
// 		return nsApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
// 	}

// 	nsApp := app.NewNameServiceApp(logger, db)

// 	return nsApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
// }


// // package main

// // import (
// // 	"fmt"
// // 	"os"
// // 	"path"

// // 	"github.com/cosmos/cosmos-sdk/client"
// // 	"github.com/cosmos/cosmos-sdk/client/flags"
// // 	"github.com/cosmos/cosmos-sdk/client/keys"
// // 	"github.com/cosmos/cosmos-sdk/client/lcd"
// // 	"github.com/cosmos/cosmos-sdk/client/rpc"
// // 	sdk "github.com/cosmos/cosmos-sdk/types"
// // 	"github.com/cosmos/cosmos-sdk/version"
// // 	"github.com/cosmos/cosmos-sdk/x/auth"
// // 	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
// // 	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
// // 	"github.com/cosmos/cosmos-sdk/x/bank"
// // 	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"

// // 	"github.com/spf13/cobra"
// // 	"github.com/spf13/viper"

// // 	"github.com/tendermint/go-amino"
// // 	"github.com/tendermint/tendermint/libs/cli"

// // 	"github.com/Sultanashazath/nameservice/app"

// // )

// // func main() {
// // 	// Configure cobra to sort commands
// // 	cobra.EnableCommandSorting = false

// // 	// Instantiate the codec for the command line application
// // 	cdc := app.MakeCodec()

// // 	// Read in the configuration file for the sdk
// // 	config := sdk.GetConfig()
// // 	config.SetBech32PrefixForAccount(sdk.Bech32PrefixAccAddr, sdk.Bech32PrefixAccPub)
// // 	config.SetBech32PrefixForValidator(sdk.Bech32PrefixValAddr, sdk.Bech32PrefixValPub)
// // 	config.SetBech32PrefixForConsensusNode(sdk.Bech32PrefixConsAddr, sdk.Bech32PrefixConsPub)
// // 	config.Seal()

// // 	// TODO: setup keybase, viper object, etc. to be passed into
// // 	// the below functions and eliminate global vars, like we do
// // 	// with the cdc

// // 	rootCmd := &cobra.Command{
// // 		Use:   "appcli",
// // 		Short: "Command line interface for interacting with appd",
// // 	}

// // 	// Add --chain-id to persistent flags and mark it required
// // 	rootCmd.PersistentFlags().String(flags.FlagChainID, "", "Chain ID of tendermint node")
// // 	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
// // 		return initConfig(rootCmd)
// // 	}

// // 	// Construct Root Command
// // 	rootCmd.AddCommand(
// // 		rpc.StatusCommand(),
// // 		client.ConfigCmd(app.DefaultCLIHome),
// // 		queryCmd(cdc),
// // 		txCmd(cdc),
// // 		flags.LineBreak,
// // 		lcd.ServeCommand(cdc, registerRoutes),
// // 		flags.LineBreak,
// // 		keys.Commands(),
// // 		flags.LineBreak,
// // 		version.Cmd,
// // 		flags.NewCompletionCmd(rootCmd, true),
// // 	)

// // 	// Add flags and prefix all env exposed with AA
// // 	executor := cli.PrepareMainCmd(rootCmd, "AA", app.DefaultCLIHome)

// // 	err := executor.Execute()
// // 	if err != nil {
// // 		fmt.Printf("Failed executing CLI command: %s, exiting...\n", err)
// // 		os.Exit(1)
// // 	}
// // }

// // func queryCmd(cdc *amino.Codec) *cobra.Command {
// // 	queryCmd := &cobra.Command{
// // 		Use:     "query",
// // 		Aliases: []string{"q"},
// // 		Short:   "Querying subcommands",
// // 	}

// // 	queryCmd.AddCommand(
// // 		authcmd.GetAccountCmd(cdc),
// // 		flags.LineBreak,
// // 		rpc.ValidatorCommand(cdc),
// // 		rpc.BlockCommand(),
// // 		authcmd.QueryTxsByEventsCmd(cdc),
// // 		authcmd.QueryTxCmd(cdc),
// // 		flags.LineBreak,
// // 	)

// // 	// add modules' query commands
// // 	app.ModuleBasics.AddQueryCommands(queryCmd, cdc)

// // 	return queryCmd
// // }

// // func txCmd(cdc *amino.Codec) *cobra.Command {
// // 	txCmd := &cobra.Command{
// // 		Use:   "tx",
// // 		Short: "Transactions subcommands",
// // 	}

// // 	txCmd.AddCommand(
// // 		bankcmd.SendTxCmd(cdc),
// // 		flags.LineBreak,
// // 		authcmd.GetSignCommand(cdc),
// // 		authcmd.GetMultiSignCommand(cdc),
// // 		flags.LineBreak,
// // 		authcmd.GetBroadcastCommand(cdc),
// // 		authcmd.GetEncodeCommand(cdc),
// // 		authcmd.GetDecodeCommand(cdc),
// // 		flags.LineBreak,
// // 	)

// // 	// add modules' tx commands
// // 	app.ModuleBasics.AddTxCommands(txCmd, cdc)

// // 	// remove auth and bank commands as they're mounted under the root tx command
// // 	var cmdsToRemove []*cobra.Command

// // 	for _, cmd := range txCmd.Commands() {
// // 		if cmd.Use == auth.ModuleName || cmd.Use == bank.ModuleName {
// // 			cmdsToRemove = append(cmdsToRemove, cmd)
// // 		}
// // 	}

// // 	txCmd.RemoveCommand(cmdsToRemove...)

// // 	return txCmd
// // }

// // // registerRoutes registers the routes from the different modules for the LCD.
// // // NOTE: details on the routes added for each module are in the module documentation
// // // NOTE: If making updates here you also need to update the test helper in client/lcd/test_helper.go
// // func registerRoutes(rs *lcd.RestServer) {
// // 	client.RegisterRoutes(rs.CliCtx, rs.Mux)
// // 	authrest.RegisterTxRoutes(rs.CliCtx, rs.Mux)
// // 	app.ModuleBasics.RegisterRESTRoutes(rs.CliCtx, rs.Mux)
// // }

// // func initConfig(cmd *cobra.Command) error {
// // 	home, err := cmd.PersistentFlags().GetString(cli.HomeFlag)
// // 	if err != nil {
// // 		return err
// // 	}

// // 	cfgFile := path.Join(home, "config", "config.toml")
// // 	if _, err := os.Stat(cfgFile); err == nil {
// // 		viper.SetConfigFile(cfgFile)

// // 		if err := viper.ReadInConfig(); err != nil {
// // 			return err
// // 		}
// // 	}
// // 	if err := viper.BindPFlag(flags.FlagChainID, cmd.PersistentFlags().Lookup(flags.FlagChainID)); err != nil {
// // 		return err
// // 	}
// // 	if err := viper.BindPFlag(cli.EncodingFlag, cmd.PersistentFlags().Lookup(cli.EncodingFlag)); err != nil {
// // 		return err
// // 	}
// // 	return viper.BindPFlag(cli.OutputFlag, cmd.PersistentFlags().Lookup(cli.OutputFlag))
// // }
