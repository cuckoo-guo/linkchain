package app

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"sort"
	"testing"
	"time"

	"github.com/lianxiangcloud/linkchain/blockchain"
    cfg "github.com/lianxiangcloud/linkchain/config"
	"github.com/lianxiangcloud/linkchain/utxo"

	"github.com/lianxiangcloud/linkchain/libs/ser"
	"github.com/lianxiangcloud/linkchain/libs/txmgr"
	"github.com/stretchr/testify/mock"

	"github.com/lianxiangcloud/linkchain/accounts/keystore"
	"github.com/lianxiangcloud/linkchain/libs/common"
	"github.com/lianxiangcloud/linkchain/libs/db"
	dbm "github.com/lianxiangcloud/linkchain/libs/db"
    "github.com/lianxiangcloud/linkchain/metrics"
	"github.com/lianxiangcloud/linkchain/state"
	"github.com/lianxiangcloud/linkchain/types"
	"github.com/stretchr/testify/assert"
)

type ks struct {
	key string
	pwd string
}

var kss = []ks{
	ks{
		key: `{"address":"54fb1c7d0f011dd63b08f85ed7b518ab82028100","crypto":{"cipher":"aes-128-ctr","ciphertext":"e77ec15da9bdec5488ce40b07a860fb5383dffce6950defeb80f6fcad4916b3a","cipherparams":{"iv":"5df504a561d39675b0f9ebcbafe5098c"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"908cd3b189fc8ceba599382cf28c772b735fb598c7dbbc59ef0772d2b851f57f"},"mac":"9bb92ffd436f5248b73a641a26ae73c0a7d673bb700064f388b2be0f35fedabd"},"id":"2e15f180-b4f1-4d9c-b401-59eeeab36c87","version":3}`,
		pwd: `1234`,
	},
	ks{
		key: `{"address":"e6a36f2e34afccdd93c8e657a9795d5d26fb3344","crypto":{"cipher":"aes-128-ctr","ciphertext":"5e759f5ddfed547733832efea4fd46d2df12c6c80430e9ab26823b3f19f2edd2","cipherparams":{"iv":"c5c54ea1db594a447afd1f0dff178345"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"5dd9ac7552cdde1dc4e0867b52d5b9d870a3c862323cbb800baca3b979100cd0"},"mac":"e48a53fbb4ca94ea5b3492acb1eca39afdd5f179bc95f13ce36f8d401fe55f4f"},"id":"ae1c927f-ebd3-45b5-88c0-d633bce79d02","version":3}`,
		pwd: "1234",
	},
}

var (
	kssmultisign = []ks{
		ks{
			key: `{"address":"3f76ec08843942fd164c66507c05bef8f8b7df70","crypto":{"cipher":"aes-128-ctr","ciphertext":"fb7ab9a926785eda97e77ef04f7496063922943236254192f28c2b7a786ceee3","cipherparams":{"iv":"4f5f25711b58361c0747122a41cf52f4"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"be0916a282b34b70a8882bbf9ec2dabbf8fe6374a3271130eadf86f715c78e82"},"mac":"84f22cb1f74adcf33463f4fdce73877e7466dbc936c93d7c0ebcf408b82bf8e9"},"id":"ff528baa-e996-48a7-9650-88c99073a8cc","version":3}`,
			pwd: `1234`,
		},
		ks{
			key: `{"address":"5f502c6a99fd83093625b54a1bf1166bdf597660","crypto":{"cipher":"aes-128-ctr","ciphertext":"c95b4b4a38f14b91d28a85aae3f6eabf1b3bdf58dabaddd43c2c387b911e3e0f","cipherparams":{"iv":"bdb2650473ad9fd3c8cd877d807c95e0"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"bbfd32589e1b2a104d0eb0fe500f341f221d10cb40006c7a548993189274b7f5"},"mac":"dd938504d8bd6358c8309d4ff1e42c2631d6a84f2e8c6dfb3853cdaab247fe2f"},"id":"3c3a15e6-77c4-49c5-b8b4-f9fe29ecfbd5","version":3}`,
			pwd: `1234`,
		},
		ks{
			key: `{"address":"599bb2d47f605b5e655609c13cdaa1450f6b73a0","crypto":{"cipher":"aes-128-ctr","ciphertext":"c04dfbbfaf5ef6b6ecaa5eae416bbe960d5b341f63cde87763ee9818f00cb6c3","cipherparams":{"iv":"8c2901a11037b8680ca1c1cfbe5878d3"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"4110345e538327bf70b52674299fb5e6264759b1a0c007406180dc4476f9e48d"},"mac":"052721103822ec1ad9eabfb975300574b2221452529f063a1cead84b3abebde5"},"id":"31bf3b76-9a4f-455a-9484-cb7cd619773e","version":3}`,
			pwd: `1234`,
		},
	}
	acc []*keystore.Key
)

var (
	accounts            []*keystore.Key
	initBalance, _      = big.NewInt(1e18).SetString("0xfffffffffffffffffffffffffff", 0)
	initTokenBalance, _ = big.NewInt(1e18).SetString("0xfffffffffffffffffffffffffff", 0)
	gasPrice            = big.NewInt(1e11)
	gasLimit            = uint64(1e5)
	zeroAddr            = common.EmptyAddress
	amount1             = big.NewInt(1)
)

func init() {
	for _, k := range kss {
		key, err := keystore.DecryptKey([]byte(k.key), k.pwd)
		if err != nil {
			panic(err)
		}
		accounts = append(accounts, key)
	}

	for _, k := range kssmultisign {
		key, err := keystore.DecryptKey([]byte(k.key), k.pwd)
		if err != nil {
			panic(err)
		}
		acc = append(acc, key)
	}
}

func getPriVals() ([]types.PrivValidator, *types.ValidatorSet) {
	validators := make([]*types.Validator, 0, 10)
	privs := make([]types.PrivValidator, 0, 10)
	for i := 0; i < 10; i++ {
		v, p := types.RandValidator(false, 1)
		validators = append(validators, v)
		privs = append(privs, p)
	}

	sort.Sort(types.ValidatorsByAddress(validators))
	vs := &types.ValidatorSet{
		Validators: validators,
	}

	if len(validators) > 0 {
		vs.IncrementAccum(1)
	}

	return privs, vs
}

func initApp() (*LinkApplication, error) {
	_, valset := getPriVals()

	stateDB := newTestStateDB()
	txpool := &MockMempool{}

	blockStore := newTestBlockStore()
	crossState := newTestCrossState(blockStore)
	//crossState := &MockCrossState{}
	blockStore.SetCrossState(crossState)

	app, err := newTestApp(stateDB, txpool, blockStore, crossState)
	if err != nil {
		return nil, err
	}
	app.SetLastChangedVals(0, valset.Validators)
	return app, nil
}

func TestApp(t *testing.T) {
	priVals, valset := getPriVals()

	stateDB := newTestStateDB()
	txpool := &MockMempool{}

	blockStore := newTestBlockStore()
	crossState := newTestCrossState(blockStore)
	//crossState := &MockCrossState{}
	blockStore.SetCrossState(crossState)

    config := cfg.DefaultConfig()
    pv := types.GenFilePV("")
    metrics.PrometheusMetricInstance.Init(config, pv.PubKey, logger.With("module", "prometheus_metrics"))

	app, err := newTestApp(stateDB, txpool, blockStore, crossState)
	if err != nil {
		t.Fatalf("initApp err:%v", err)
	}
	app.SetLastChangedVals(0, valset.Validators)
	txs := make(types.Txs, 0, 3000)
	for i := 0; i < 1; i++ {
		nonce := app.checkTxState.GetNonce(accounts[0].Address)
		balance := app.checkTxState.GetBalance(accounts[0].Address)
		fmt.Println("###### nonce", nonce, balance)
		to := common.HexToAddress(fmt.Sprintf("0x%x", i%64))
		tx, _ := genTx(accounts[0], nonce, &to, big.NewInt(1), nil)
		txs = append(txs, tx)
		fmt.Printf("%s: %s \n", tx.TypeName(), tx.Hash().Hex())
		if err := app.CheckTx(tx, false); err != nil {
			t.Fatalf("CheckTx err:%v", err)
		}
	}

	for i := 0; i < 1; i++ {
		nonce := app.checkTxState.GetNonce(accounts[0].Address)
		to := common.HexToAddress(fmt.Sprintf("0x%x", i%64))
		tx, _ := genTokenTx(accounts[0], &to, zeroAddr, nonce, big.NewInt(1), 0, "")
		txs = append(txs, tx)
		fmt.Printf("%s: %s\n", tx.TypeName(), tx.Hash().Hex())
		if err := app.CheckTx(tx, false); err != nil {
			t.Fatalf("CheckTx err:%v", err)
		}
	}

	//types.TxUpdateValidatorsType,
	//types.TxContractCreateType,
	for i := 0; i < 2; i++ {
		mstAddr := common.BytesToAddress([]byte("mst"))
		nonce := app.checkTxState.GetNonce(mstAddr)
		tx := genMultiSignAccountTx(nonce, types.SupportType(i))

		signBytes, err := types.GenMultiSignBytes(tx.MultiSignMainInfo)
		if err != nil {
			t.Error(" GenMultiSignBytes error!!")
		}

		sigs := make([]types.ValidatorSign, 0)
		for _, v := range priVals {
			signature, _ := v.SignData(signBytes)
			sig := types.ValidatorSign{
				Addr:      v.GetAddress().Bytes(),
				Signature: signature,
			}
			sigs = append(sigs, sig)
		}
		tx.Signatures = sigs

		if err := tx.VerifySign(valset); err != nil {
			t.Error(err)
		}

		txs = append(txs, tx)
		fmt.Printf("%s: %s\n", tx.TypeName(), tx.Hash().Hex())
		if err := app.CheckTx(tx, false); err != nil {
			t.Fatalf("CheckTx err:%v nonce %d", err, nonce)
		}
	}
	txpool.On("VerifyTxFromCache", mock.Anything).Return(nil, false)
	txpool.On("Lock").Return()
	txpool.On("Unlock").Return()
	txpool.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	txpool.On("Reap", mock.Anything).Return(txs)
	txpool.On("KeyImageReset").Return()
    txpool.On("GetTxFromCache", mock.Anything).Return(nil)

	height := uint64(1)
	block := app.CreateBlock(height, 1000, 1e18, uint64(time.Now().Unix()))
	block.LastCommit = &types.Commit{}
    app.PreRunBlock(block)
	if !app.CheckBlock(block) {
		t.Fatalf("CheckBlock not ok")
	}

	partSet := block.MakePartSet(types.DefaultConsensusParams().BlockGossip.BlockPartSizeBytes)
	_, err = app.CommitBlock(block, partSet, &types.Commit{}, false)
	if err != nil {
		t.Fatalf("CommitBlock err:%v", err)
	}

	types.RegisterBlockAmino()

	js, _ := json.Marshal(app.lastTxsResult)
	fmt.Println("lastTxsResult:", string(js))

	for _, tx := range txs {
		_, txEntry := blockStore.GetTx(tx.Hash())
		js, _ = json.Marshal(txEntry)
		fmt.Println("txEntry:", tx.Hash().Hex(), string(js))
	}

	txs = nil
	nonce := app.checkTxState.GetNonce(accounts[0].Address)
	gasLimit := uint64(0)
    demoTokenBin := "60806040526012600160006101000a81548160ff021916908360ff16021790555034801561002c57600080fd5b50e4801561003957600080fd5b50600160009054906101000a900460ff1660ff16600a0a61271002600081905550600054e07fd1398bee19313d6bf672ccb116e51f4a1a947e91c757907f51fbb5b5e56c698f60003030600054604051808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200194505050505060405180910390a16107eb806101436000396000f300608060405260043610610083576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806306fdde0314610085578063313ce567146101225780633eaaf86b146101605780635d0268e61461019857806370a08231146101b8578063a4556fce1461021c578063d0ca623414610226575b005b34801561009157600080fd5b50e4801561009e57600080fd5b506100a7610230565b6040518080602001828103825283818151815260200191508051906020019080838360005b838110156100e75780820151818401526020810190506100cc565b50505050905090810190601f1680156101145780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561012e57600080fd5b50e4801561013b57600080fd5b5061014461026d565b604051808260ff1660ff16815260200191505060405180910390f35b34801561016c57600080fd5b50e4801561017957600080fd5b50610182610280565b6040518082815260200191505060405180910390f35b6101b660048036038101908080359060200190929190505050610286565b005b3480156101c457600080fd5b50e480156101d157600080fd5b50610206600480360381019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506103b0565b6040518082815260200191505060405180910390f35b6102246103e8565b005b61022e6105cf565b005b60606040805190810160405280600981526020017f44656d6f546f6b656e0000000000000000000000000000000000000000000000815250905090565b600160009054906101000a900460ff1681565b60005481565b3073ffffffffffffffffffffffffffffffffffffffff16e273ffffffffffffffffffffffffffffffffffffffff161415156102c057600080fd5b80e41480156102cf5750600081115b15156102da57600080fd5b7fd1398bee19313d6bf672ccb116e51f4a1a947e91c757907f51fbb5b5e56c698f3330e2e4604051808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200194505050505060405180910390a150565b60008173ffffffffffffffffffffffffffffffffffffffff163073ffffffffffffffffffffffffffffffffffffffff16e19050919050565b6000e41115156103f757600080fd5b3373ffffffffffffffffffffffffffffffffffffffff163073ffffffffffffffffffffffffffffffffffffffff16e4e37fd1398bee19313d6bf672ccb116e51f4a1a947e91c757907f51fbb5b5e56c698f3330e2e4604051808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200194505050505060405180910390a17fd1398bee19313d6bf672ccb116e51f4a1a947e91c757907f51fbb5b5e56c698f303330e4604051808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200194505050505060405180910390a1565b600080341115156105df57600080fd5b6002340290503373ffffffffffffffffffffffffffffffffffffffff163073ffffffffffffffffffffffffffffffffffffffff1682e37fd1398bee19313d6bf672ccb116e51f4a1a947e91c757907f51fbb5b5e56c698f3330600034604051808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200194505050505060405180910390a17fd1398bee19313d6bf672ccb116e51f4a1a947e91c757907f51fbb5b5e56c698f30333084604051808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200194505050505060405180910390a1505600a165627a7a7230582035535fe5dabdc379daafa902bd1a8b41026cc07fe489cad1a962214b964671dd0029"
	ctx := genContractCreateTx(accounts[0].Address, gasLimit, nonce, demoTokenBin)
    assert.NotNil(t, ctx)
	if err := app.CheckTx(ctx, false); err != nil {
		t.Fatalf("CheckTx err:%v", err)
	} else {
		t.Logf("CheckTx err:%v", err)
	}
	fmt.Println("ctx:", ctx.String())
	txs = append(txs, ctx)
	txpool = &MockMempool{}
	txpool.On("VerifyTxFromCache", mock.Anything).Return(nil, false)
	txpool.On("Lock").Return()
	txpool.On("Unlock").Return()
	txpool.On("Update", mock.Anything, mock.Anything).Return(nil)
	txpool.On("Reap", mock.Anything).Return(txs)
    txpool.On("GetTxFromCache", mock.Anything).Return(nil)
	txpool.On("KeyImageReset").Return()
	app.SetMempool(txpool)
	//block2
	height = uint64(2)
	block = app.CreateBlock(height, 1000, 1e18, uint64(time.Now().Unix()))
	block.LastCommit = &types.Commit{}
    //app.PreRunBlock(block)
	if !app.CheckBlock(block) {
		t.Logf("CheckBlock not ok")
	}

	partSet = block.MakePartSet(types.DefaultConsensusParams().BlockGossip.BlockPartSizeBytes)
	_, err = app.CommitBlock(block, partSet, &types.Commit{}, false)
	if err != nil {
		t.Fatalf("CommitBlock err:%v", err)
	}

	for _, tx := range txs {
		_, txEntry := blockStore.GetTx(tx.Hash())
		js, _ = json.Marshal(txEntry)
		fmt.Println("txEntry:", tx.Hash().Hex(), string(js))
		receipt, _, _, _ := blockStore.GetTransactionReceipt(tx.Hash())
		rc, _ := json.Marshal(receipt)
		fmt.Println("receipt:", string(rc))
	}

}

func createTokenTrans(from *keystore.Key, to *common.Address, tokenAddress common.Address, nonce uint64, amount *big.Int, ret uint8, reterr string) (*types.TokenTransaction, error) {
	signedTx, err := genTokenTx(from, to, tokenAddress, nonce, amount, ret, reterr)
	if err != nil {
		panic(err)
	}
	return signedTx, err
}

func genTokenTx(from *keystore.Key, to *common.Address, tokenAddress common.Address, nonce uint64, amount *big.Int, ret uint8, err string) (*types.TokenTransaction, error) {
	var tx *types.TokenTransaction
	tx = types.NewTokenTransaction(tokenAddress, nonce, *to, amount, gasLimit, gasPrice, []byte(""))
	if err := tx.Sign(types.GlobalSTDSigner, from.PrivateKey); err != nil {
		return nil, err
	}
	return tx, nil
}

func genTx(from *keystore.Key, nonce uint64, to *common.Address, amount *big.Int, payload []byte) (*types.Transaction, error) {
	toAddr := common.EmptyAddress
	if to != nil {
		toAddr = *to
	}
	gasLimit := types.CalNewAmountGas(amount, types.EverLiankeFee)
	tx := types.NewTransaction(nonce, toAddr, amount, gasLimit, gasPrice, payload)
	if err := tx.Sign(types.GlobalSTDSigner, from.PrivateKey); err != nil {
		return nil, err
	}
	return tx, nil
}

func genContractTx(from *keystore.Key, nonce uint64, to *common.Address, amount *big.Int, payload []byte) (*types.Transaction, error) {
	tx := types.NewContractCreation(nonce, amount, gasLimit, gasPrice, payload)
	if err := tx.Sign(types.GlobalSTDSigner, from.PrivateKey); err != nil {
		return nil, err
	}
	return tx, nil
}

func genTxForCreateContract(from *keystore.Key, gas uint64, nonce uint64, contractFile string) *types.Transaction {
	bin, err := ioutil.ReadFile(contractFile)
	if err != nil {
		panic(err)
	}
	ccode := common.Hex2Bytes(string(bin))
	gasLimit = gas
	tx, _ := genContractTx(from, nonce, nil, big.NewInt(0), ccode)
	return tx
}

func genContractCreateTx(fromaddr common.Address, gasLimit uint64, nonce uint64, contractFile string) *types.ContractCreateTx {
	//gasPrice := big.NewInt(1e11)
	var ccode []byte
	if len(contractFile) < 100 {
		bin, err := ioutil.ReadFile(contractFile)
		if err != nil {
			panic(err)
		}
		ccode = common.Hex2Bytes(string(bin))
	} else {
		ccode = common.Hex2Bytes(contractFile)
	}

	ccMainInfo := &types.ContractCreateMainInfo{
		FromAddr:     fromaddr,
		AccountNonce: nonce,
		Amount:       big.NewInt(0),
		Payload:      ccode,
		//GasLimit:     gasLimit,
		//Price:        gasPrice,
	}
    signatures := make([][]byte, 0)
    for i := 0; i < 2; i++ {
        priveKey := acc[i].PrivateKey
        sigData, err := types.SignContractCreateTx(priveKey, ccMainInfo)
        if err != nil {
            fmt.Printf("SignContractCreateTx failed:%v", err)
            return nil
        }
        signatures = append(signatures, sigData)
    }
    priveKey := accounts[0].PrivateKey
    sigData, _ := types.SignContractCreateTx(priveKey, ccMainInfo)
    signatures = append(signatures, sigData)

	tx := types.CreateContractTx(ccMainInfo, signatures)
	return tx
}

func genContractUpgradeTx(fromaddr common.Address, contract common.Address, nonce uint64, contractFile string) *types.ContractUpgradeTx {
	bin, err := ioutil.ReadFile(contractFile)
	if err != nil {
		panic(err)
	}
	ccode := common.Hex2Bytes(string(bin))

	ccMainInfo := &types.ContractUpgradeMainInfo{
		FromAddr:     fromaddr,
		Recipient:    contract,
		AccountNonce: nonce,
		Payload:      ccode,
	}
	tx := types.UpgradeContractTx(ccMainInfo, nil)
	return tx
}

func genMultiSignAccountTx(nonce uint64, supportType types.SupportType) *types.MultiSignAccountTx {
	return &types.MultiSignAccountTx{
		MultiSignMainInfo: types.MultiSignMainInfo{
			AccountNonce:  nonce,
			SupportTxType: supportType,
			SignersInfo: types.SignersInfo{
				MinSignerPower: 20,
				Signers: []*types.SignerEntry{
					&types.SignerEntry{
						Power: 10,
						Addr:  acc[0].Address,
					},
					&types.SignerEntry{
						Power: 10,
						Addr:  acc[1].Address,
					},
					&types.SignerEntry{
						Power: 10,
						Addr:  acc[2].Address,
					},
				},
			},
		},
	}
}

func newTestCrossState(blockStore *blockchain.BlockStore) *txmgr.Service {
	return txmgr.NewCrossState(db.NewMemDB(), blockStore)
}

func newTestStateDB() dbm.DB {
	return dbm.NewMemDB()
}

func newTestBlockStore() *blockchain.BlockStore {
	blockStoreDB := db.NewMemDB()
	return blockchain.NewBlockStore(blockStoreDB)
}

func newTestApp(sdb dbm.DB, txpool types.Mempool, blockStore *blockchain.BlockStore, crossState *txmgr.Service) (*LinkApplication, error) {
	block := &types.Block{
		Header: &types.Header{
			Height:     0,
			Coinbase:   common.HexToAddress("0x0000000000000000000000000000000000000000"),
			Time:       uint64(1507737600),
			NumTxs:     0,
			TotalTxs:   0,
			ParentHash: common.EmptyHash,
			StateHash:  common.EmptyHash,
			GasLimit:   types.DefaultConsensusParams().BlockSize.MaxGas,
		},
		Data:       &types.Data{},
		LastCommit: &types.Commit{},
	}

	txsResult := types.TxsResult{}

	BlockPartSet := types.DefaultConsensusParams().BlockGossip.BlockPartSizeBytes
	blockParts := block.MakePartSet(BlockPartSet)

	blockStore.SaveBlock(block, blockParts, nil, nil, &txsResult)

	utxoStore := utxo.NewUtxoStore(dbm.NewMemDB(), dbm.NewMemDB(), dbm.NewMemDB())
	utxoStore.SetLogger(logger.With("module", "apptest"))

	//var linkApp *LinkApplication
	balanceRecord := blockchain.NewBalanceRecordStore(dbm.NewMemDB(), false)
    eventBus := types.NewEventBus()
    eventBus.SetLogger(logger.With("module", "events"))
	linkApp, err := NewLinkApplication(sdb, blockStore, utxoStore, crossState, eventBus, false, balanceRecord, nil, nil)
	linkApp.SetMempool(txpool)
    linkApp.SetLogger(logger.With("module", "apptest"))
	for i := 0; i < 2; i++ {
		state := linkApp.storeState
		if i == 1 {
			state = linkApp.checkTxState
		}
		for _, acc := range accounts {
			state.AddBalance(acc.Address, initBalance)
			state.AddTokenBalance(acc.Address, tokenAddr, initTokenBalance)
		}
		for _, acc := range acc {
			state.AddBalance(acc.Address, initBalance)
			state.AddTokenBalance(acc.Address, tokenAddr, initTokenBalance)
		}
	}

	return linkApp, err
}

func newTestState() *state.StateDB {
	sdb := dbm.NewMemDB()
	state, _ := state.New(common.EmptyHash, state.NewDatabase(sdb))

	for _, acc := range accounts {
		state.AddBalance(acc.Address, initBalance)
		state.AddTokenBalance(acc.Address, tokenAddr, initTokenBalance)
		state.IntermediateRoot(false)
	}
	return state
}

func init() {
	types.RegisterUTXOTxData()
}

func genUTXOTransaction(hextx string) *types.UTXOTransaction {
	var utxoTx types.UTXOTransaction
	hexData, err := hex.DecodeString(hextx)
	if err != nil {
		fmt.Printf("hex Decode err %v\n", err)
		return nil
	}
	err = ser.DecodeBytes(hexData, &utxoTx)
	if err != nil {
		fmt.Printf("DecodeBytes err %v\n", err)
		return nil
	}
	fmt.Printf("UTXOTx\n: %s\n", utxoTx)
	return &utxoTx
}
