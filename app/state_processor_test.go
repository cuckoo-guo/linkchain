package app

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/lianxiangcloud/linkchain/accounts/abi"
	"github.com/lianxiangcloud/linkchain/libs/common"
	"github.com/lianxiangcloud/linkchain/libs/crypto"
	lctypes "github.com/lianxiangcloud/linkchain/libs/cryptonote/types"
	//"github.com/lianxiangcloud/linkchain/libs/hexutil"
	"github.com/lianxiangcloud/linkchain/libs/log"
	"github.com/lianxiangcloud/linkchain/state"
	"github.com/lianxiangcloud/linkchain/types"
	"github.com/lianxiangcloud/linkchain/vm/evm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	blocksNum     = 10
	txNumPerBlock = 10
	coinbase      = common.HexToAddress("0x0")
	testToAddr    = common.HexToAddress("0x3")
	tokenAddr     = common.HexToAddress("0x37c9b94a0f4816ff9e209ff9fe56e2a094deefd7")
	logger        = log.Root()
	gasUsed       = new(big.Int).Mul(new(big.Int).SetUint64(gasLimit), gasPrice)
)

func init() {

}

func TestProcess(t *testing.T) {
	loopSum := blocksNum
	initBalance = new(big.Int).Mul(big.NewInt(1e18), big.NewInt(1000000))
	initTokenBalance = new(big.Int).Mul(big.NewInt(1e18), big.NewInt(50))
	st := newTestState()
	states := make([]*state.StateDB, loopSum)

	for i := range states {
		states[i] = st.Copy()
	}

	app, err := initApp()
	if err != nil {
		log.Error("initApp", "err", err)
		return
	}
	sp := NewStateProcessor(nil, app)
	vc := evm.Config{EnablePreimageRecording: false}
	blocks := make([]*types.Block, loopSum)

	var (
		i            uint64
		bfBalance    *big.Int
		afBalance    *big.Int
		//toBalance    *big.Int
		//toBalance2   *big.Int
		transferGas  uint64
		//txFee        *big.Int
		receipts     types.Receipts
		blockGas     uint64
		//utxoOutputs  []*types.UTXOOutputData
		//keyImages    []*lctypes.Key
		amount       *big.Int
		//toAddr       common.Address
		//tx           *types.UTXOTransaction
	)
	//account -> account transfer
    /*
	i = 0
	//{"from":"0x54fb1c7d0f011dd63b08f85ed7b518ab82028100","nonce":"0x0","dests":[{"addr":"0xa73810e519e1075010678d706533486d8ecc8000","amount":"0x56bc75e2d63100000","data":""}]}
	toAddr = common.HexToAddress("0xa73810e519e1075010678d706533486d8ecc8000")
	toBalance = states[i].GetBalance(toAddr)
	bfBalance = states[i].GetBalance(accounts[0].Address)
	blocks[i] = genBlockAccount2Account(i, states[i])
	receipts, _, blockGas, _, utxoOutputs, keyImages, err = sp.Process(blocks[i], states[i], vc)
	require.Nil(t, err)
	amount, _ = hexutil.DecodeBig("0x56bc75e2d63100000")
	tx = blocks[i].Data.Txs[0].(*types.UTXOTransaction)
	transferGas = types.CalNewAmountGas(amount, types.EverLiankeFee)
	txFee = tx.Fee
	assert.Equal(t, transferGas, big.NewInt(0).Div(tx.Fee, big.NewInt(types.ParGasPrice)).Uint64())
	assert.Equal(t, bfBalance.Sub(bfBalance, big.NewInt(0).Add(amount, txFee)), states[i].GetBalance(accounts[0].Address))
	toBalance2 = states[i].GetBalance(toAddr)
	assert.Equal(t, amount, toBalance2.Sub(toBalance2, toBalance))
	assert.Equal(t, uint64(1), states[i].GetNonce(accounts[0].Address))
	assert.Equal(t, 1, len(receipts))
	assert.Equal(t, transferGas, blockGas)
	assert.Equal(t, transferGas, receipts[0].GasUsed)
	assert.Equal(t, 0, len(utxoOutputs))
	assert.Equal(t, 0, len(keyImages))

	//account->contract
	//{"from":"0x54fb1c7d0f011dd63b08f85ed7b518ab82028100","nonce":"0x1","dests":[{"addr":"0x37c9b94a0f4816ff9e209ff9fe56e2a094deefd7","amount":"0x56bc75e2d63100000","data":"0xd0ca6234"}
	i = 1
	toAddr = common.HexToAddress("0x37c9b94a0f4816ff9e209ff9fe56e2a094deefd7")
	toBalance = states[i].GetBalance(toAddr)
	bfBalance = states[i].GetBalance(accounts[0].Address)
	blocks[i] = genBlockAccountToContract(i, states[i])
	receipts, _, blockGas, _, utxoOutputs, keyImages, err = sp.Process(blocks[i], states[i], vc)
	require.Nil(t, err)
	amount, _ = hexutil.DecodeBig("0x56bc75e2d63100000")
	tx = blocks[i].Data.Txs[1].(*types.UTXOTransaction)
	transferGas = types.CalNewAmountGas(amount, types.EverLiankeFee)
	txFee = tx.Fee
	assert.True(t, transferGas < big.NewInt(0).Div(txFee, big.NewInt(types.ParGasPrice)).Uint64())
	assert.Equal(t, bfBalance.Sub(bfBalance, big.NewInt(0).Add(amount, txFee)), states[i].GetBalance(accounts[0].Address))
	toBalance2 = states[i].GetBalance(toAddr)
	assert.Equal(t, amount, toBalance2.Sub(toBalance2, toBalance))
	assert.Equal(t, uint64(2), states[i].GetNonce(accounts[0].Address))
	assert.Equal(t, 2, len(receipts))
	assert.Equal(t, big.NewInt(0).Div(txFee, big.NewInt(types.ParGasPrice)).Uint64(), blockGas)
	assert.Equal(t, uint64(0), receipts[0].GasUsed)
	assert.Equal(t, big.NewInt(0).Div(txFee, big.NewInt(types.ParGasPrice)).Uint64(), receipts[1].GasUsed)
	assert.Equal(t, 0, len(utxoOutputs))
	assert.Equal(t, 0, len(keyImages))

	i = 2
	//UTXO->contract + UTXOExchange
	//from = accounts[0]
	//{"subaddrs":[0],"dests":[{"addr":"0x37c9b94a0f4816ff9e209ff9fe56e2a094deefd7","amount":"0x12a05f200","data":"0xd0ca6234"}]}
	toAddr = common.HexToAddress("0x37c9b94a0f4816ff9e209ff9fe56e2a094deefd7")
	toBalance = states[i].GetBalance(toAddr)
	bfBalance = states[i].GetBalance(accounts[0].Address)
	blocks[i] = genBlockUTXOCallContract(i, states[i])
	receipts, _, blockGas, _, utxoOutputs, keyImages, err = sp.Process(blocks[i], states[i], vc)
	require.Nil(t, err)
	amount, _ = hexutil.DecodeBig("0x12a05f200")
	tx = blocks[i].Data.Txs[1].(*types.UTXOTransaction)
	transferGas = uint64(0)
	txFee = tx.Fee
	assert.True(t, transferGas < big.NewInt(0).Div(txFee, big.NewInt(types.ParGasPrice)).Uint64())
	assert.Equal(t, bfBalance.Sub(bfBalance, big.NewInt(0).Add(amount, txFee)), states[i].GetBalance(accounts[0].Address))
	toBalance2 = states[i].GetBalance(toAddr)
	assert.Equal(t, amount, toBalance2.Sub(toBalance2, toBalance))
	assert.Equal(t, uint64(1), states[i].GetNonce(accounts[0].Address))
	assert.Equal(t, 2, len(receipts))
	assert.Equal(t, big.NewInt(0).Div(txFee, big.NewInt(types.ParGasPrice)).Uint64(), blockGas)
	assert.Equal(t, uint64(0), receipts[0].GasUsed)
	assert.Equal(t, big.NewInt(0).Div(txFee, big.NewInt(types.ParGasPrice)).Uint64(), receipts[1].GasUsed)
	assert.Equal(t, 1, len(keyImages))
	assert.Equal(t, 1, len(utxoOutputs))
    */

	i = 3
	//accounts[0]->subaddr[0]  =>  accounts[0]->subaddr[1]
	blocks[i] = genBlockUTXO2UTXO(i, states[i], big.NewInt(0).Mul(types.DefaultCoefficient().UTXOFee, big.NewInt(types.ParGasPrice)))
	receipts, _, blockGas, _, _, _, err = sp.Process(blocks[i], states[i], vc)
	require.Nil(t, err)

    blocks[i] = genBlockUTXO2UTXO(i, states[i], big.NewInt(0))
	receipts, _, blockGas, _, _, _, err = sp.Process(blocks[i], states[i], vc)
	require.NotNil(t, err)

	i = 4
	bfBalance = states[i].GetBalance(accounts[0].Address)
	testToAddr = common.HexToAddress("0x3")
	amount = big.NewInt(1e18)
	transferGas = types.CalNewAmountGas(amount, types.EverLiankeFee)
	transferFee := big.NewInt(0).Mul(big.NewInt(0).SetUint64(transferGas), big.NewInt(types.ParGasPrice))

	blocks[i] = genBlockWithLocalTransaction(i)
	_, _, blockGas, _, _, _, err = sp.Process(blocks[i], states[i], vc)
	require.Nil(t, err)
	afBalance = states[i].GetBalance(accounts[0].Address)
	assert.Equal(t, big.NewInt(0).Add(amount, transferFee), big.NewInt(0).Sub(bfBalance, afBalance))
	assert.Equal(t, amount, states[i].GetBalance(testToAddr))
	assert.Equal(t, transferGas, blockGas)

	i = 5
	bfBalance = states[i].GetBalance(accounts[0].Address)
    bfToken := states[i].GetTokenBalance(accounts[0].Address, tokenAddr)
    fmt.Println("bfToken", "account", accounts[0].Address.String(), "tokenBalance", bfToken)
	blocks[i] = genBlockUTXOTokenTransaction(i, states[i])
	receipts, _, blockGas, _, _, _, err = sp.Process(blocks[i], states[i], vc)
	afBalance = states[i].GetBalance(accounts[0].Address)
	require.Nil(t, err)
    transferGas = types.CalNewAmountGas(amount, types.EverLiankeFee) + 600000
    transferMoneyInGas := 1e7 - transferGas
	recvToken := big.NewInt(0).Mul(big.NewInt(int64(transferMoneyInGas)*2), big.NewInt(types.ParGasPrice))
	contractAddr := tokenAddr
	assert.Equal(t, big.NewInt(0).Add(recvToken, bfToken), states[i].GetTokenBalance(accounts[0].Address, tokenAddr))
	conTractToken := big.NewInt(0).Mul(big.NewInt(10000), big.NewInt(1e18))
	assert.Equal(t, big.NewInt(0).Sub(conTractToken, recvToken), states[i].GetTokenBalance(contractAddr, tokenAddr))
	assert.Equal(t, 3, len(receipts))
	assert.Equal(t, uint64(70233+types.MinGasLimit+types.MinGasLimit+types.MinGasLimit), blockGas) // 2TxGas + MinGas + MinGas
	assert.Equal(t, afBalance, big.NewInt(0).Sub(bfBalance, big.NewInt(0).SetUint64(blockGas*1e11+transferMoneyInGas*1e11)))
}

func genBlockUTXO2Account(height uint64, statedb *state.StateDB) *types.Block {
	fmt.Println("account[0].address", accounts[0].Address.String(), "account[1].address", accounts[1].Address.String())

	txs := make(types.Txs, 0)
	var serTxUTXOToAccountOnlyTransferValue string = "f90b3d80f84a10c698b1d37308f84180de8202308201768203448201c28201bb82033481d382016c82059582020729a04ba2bcea3435c6becdaf044f578faadadc3086359aac91850c347bcc99d73644f86fc77cca059e4aa7f83c947b6837189a3464d3c696069b2b42a9ae8e17dda18405f5e10080a0b84f4ba682f169b3b91e7ae28d38daf6640b7ad7c703f026bea1975ee510f1a45842699137a517e2a0ad852eec1151adcff0db94cad0ed8def22b9e85afa1994e159cd26646550d40f809400000000000000000000000000000000000000009454fb1c7d0f011dd63b08f85ed7b518ab82028100a061c8f12510c990695497e8db181a2a60c5670ff6c35d9ee61137b108db7270228405f5e10080f84582ec8da005376bfa0f1c247315bf7153f9e184451ca0293e0fc49f8fc7df04f5732fdb86a053d7575c6faaae60e681cd8d85b913e4b39d052bfa911ed331ef566374ed1499f909e4f903c303a058591aa39a11ff3408419044190d1f1c78e8870e9fb4f46e2167431dd96527ecf902eff902ecf842a0246ac43d9a885e03355830c638a8eda22c93be7128364918cf9cf9beadf1e5c6a0d8fc7fd88c43a1b296cbd26ddbe95bbd0a06867b0400df631f281a0c468b4984f842a088a1bbfba3d9576b310cd746772d7c7baa9f3a0244ee1f0504478262ec73b942a0e69e74e0ba302d3c2f5ab496d2ac1880077d4ea3e36584453acbd5c5dabb04cff842a0f9b59b309733938e124ebd035014f06ecd8b7caaf9a35fac6c168bdc55569feaa088506386a7329f51448a4ff48009e7a4016578610c0311ec1e499d690839fee9f842a09dd2161f837016abaeb4c1483c6d251a025b8e8a474fcda03bcb37cc62cae163a0a9435f034a0d805f27fbc9050166401ff38afb0510d629dbf9c7206b39577746f842a0713608819469ea83bfa9d82ab4d4e0628f4302a70502d865367380c5114fd684a0074e3bad7396553ae860289c1a5a6d1cbe92154a701e064255a17af48575d51af842a0261d6d3cdbc606a12c9eddb673aaf8ad8fed78e0121dbcdc2f3af0e923d6c419a0d1e384986f3eadaeefc09dbb9e707c1c184fc71a3b7f31e8c6d56905e40f705cf842a0f7e132ebff6ad453be59ab73e14605d0b568f3a458131367b0783e455cf678a9a030f6d6bb3a397ac26a3f33e797d992f4b63dc6a34d23839d76b150da76fc7858f842a0bbce3a3462a0943411f175bae25c4d3eaeedeab2c0f91775eef2bad781fea5c3a03af0386c66012383b3fab2b2f78d1810ad7816906ee2944cfbdcccf626047710f842a0703ea122b7ab7597748a1dfb9098d044682403cc6628224d2282b2cd738c957ca03c5f2d350c01a39b684ab87c1046ead63381876414c38c4b7d7b80c3af169e97f842a01b5a616588c1fcda21fc154150ebe7ebc54c51863037f3caaa65f59bc1e482f0a062d8777665e1d1f9486b54e2640ecd7ccfa700797e04f8793717e6f26b45a9c0f842a01bad937874d1168fcd8acd9977e49ee86895c3b693f0b47bb26eaa634eb57182a06ba90daa26fd3cc766a4af691aadd16cf765b15974b2007bf4a8e97f4ad82118c0f865f863a0f0229c9dee191ca104d403e708010f71959c15582e89fef9a0dab0d13aaf9f0ea0896a27a47591ff726faa1aa4e436231e91b6b815ec76f1816d516d3e9da37e00a00000000000000000000000000000000000000000000000000000000000000000f844f842a00000000000000000000000000000000000000000000000000000000000000000a0232ec65bf43a4b5ab436988fc82ac55075835cf88fce1418e08b9bfa343f706680f9061bc0f902bdf902bac0a0a33e9cdfec0da5cdfac403a303fcdebb603fa755980c1998d4e3de8378f39806a040dec11d61571deda397118a45c255c445e3fa9c196f0174a8071f8fcbfb2ceba0a7d2f4303ce4354ee36d115cf57362278461719b88510a6ab4cf2e89964c2e74a0190e254edc3bedb057c75575607c7145230b92258c2760d7e9bff80ae675af98a0be5ad88fb50773cd8debcc200343ab2db202d88eb6accdcfbc445ff3b381280ba023199532e2fde0f17fb9bb800bb579b5cca88bfa1139eb7e3fa937aef1fda900f8c6a0747a45d5709909d9eb44a988939140f742d769f5a1ff1f55506b81153f79509ca0a5aba63f008a8889ba8b10b8acf497fb042691421fa6e85cc41c01c1e8086c83a0f070d9e8df9200435a99392f99747eacc3870ce64cd4981d765304e5e3afb8c8a066a75d8aa2968abd09377bb8f65c8233e51ed896ff717710b59d2f365c0e8d60a087de4aa1e29be43fe419475bba4d46d8ca1d6d73a882d1e8c8a922e34bded996a04e73b426afc8e19265c4fcde818eea99c676a82ca392298d0e7895a07d6cb956f8c6a0aa5746afbb4b9dde55786a694a9714146782e47c724443ada3e9b23a581e7822a02efb62c2d7babbd6da7cc55963fd837fd59ff73a24090ca48a53aa8964ae8a6ca0cc388d8c0a3bb84b66faa8715a39eb9481640ef6512e297f87139775b290e24aa0812937f8b9797ba3dcfae818e888bf12ec5b13ce53dc6e433a2e78ab2d5f81c3a04f0c37e9c54a27d9d75ff8650acca2a252e03e22679102cf130e053df2a500c6a015ecb49eacef8e86e57e9a1bce0600911521c4189582e63ca2ec60c74f450faba03d15829f956b2dda5b675f56acd9c1315f569479a1a42eceaf585e549d9a700ba08371039b5692d00fc0d32cd7ee1309e980f4ffaa115f76ecd6eadac03b68b504a0fe546b55b67ad1d73a041767d654633b8bd1b0f67945418bedededed6c44e000f90335f90332f902ecf842a018da924ae68dff868b8a6a9aa7dae0ec801b983ec270a08a59a29b510d234906a07d8c68f7647f7ab8c05939d2d6cf3b725cc4f56fc9a247b4c292e86a0fa2b903f842a0b7e891ac0c51e1412413e779714bac7e192c873877a382a1bfc4b6bd299ed407a08600be0e8c10d9816662b07201457e2a1e76d54491bcf5ea21274abce1c0ec07f842a07e47801300537bebf45ee7145cbfc76d08a14773d8987cf99b60521fbd51250aa0b69b68806a08a193686d5a316d946188c5b140d378dd4c72aa0ef7c971f10906f842a0e5d0d2183f99767aedcb2d13544ba397ceae704da3c6bdd9564a4427d3830603a0238ec631b76c736feac0a799c3a111ab4ffcaf47ec4ef5f50817359841d9f506f842a0569aa38070fa6a35b344fc51157eadc106d4b240b62b1b1780c46d0819006f09a02a97dabf797c3b29f9e33130e58d991003706b27bb138e2a12491a85aa59410ff842a016aa308ade52e4a72fc21fb53a745fb93c7b800e6916249b45d81e41baad0705a0ab08073b6a04b71a78a6ed7aa1f150ec7660bee82558a189e403ed50e63ea508f842a096381444bb3ace25d4be114f64a00629ce17ff1ef1e8fcd4a03e109d292b830aa0c022397061b84105413cee30a492cea8b74592659af2f5c8ec3e864da5e29009f842a073c4be6c8dbc4fb1cff31fd48ccbe456c58627b9dd31d0593a77aefe0f7a5208a0b4623b66862159732dca641ce261a70d3782999ccdef33a7bc8e140952347600f842a071ba79e33bc907e98455c688c352dd63ede3f3a395e76e709cb6e46ea3ad100da0bfc8c9d4572bbe2c20e68c1628bbd2bed3d9943ed116851d06f05d00ca4a0d08f842a0b5767e45ac8622280458277b5461d2ccdffd75676a16604c9f4ac948a5366d08a0bff2eea7e3f2a502b032279a515c52a4b6c47e87dadecc402760c1ce1dc6a80bf842a05c81ec5f380b457e844b509b6508ca6625213dd5c7b2baa8231e2def8e676600a05cb91a4100993d0b15472ea672f3be71117349e9cfdc42163ab0fcf7ffd3e909a0a27163cf17430323a67d93015ef2837a2db3b74d0008167bf59c46839d75b30ae1a04ba2bcea3435c6becdaf044f578faadadc3086359aac91850c347bcc99d73644e1a00d8178b01b8b6705f2cfb1eede5a2e059d3604b58fe33a1de4d3af56a176860b"
	// accountOutput.Amount = 1e8  Fee=1e8
	utxoTx := genUTXOTransaction(serTxUTXOToAccountOnlyTransferValue)

	txs = append(txs, utxoTx)
	/*
	   if err := utxoTx.CheckBasic(nil); err != nil {
	       log.Error("CheckBasic", "err", err)
	   }
	*/
	block := &types.Block{
		Header: &types.Header{
			Height:     height,
			Coinbase:   coinbase,
			Time:       uint64(time.Now().Unix()),
			NumTxs:     uint64(len(txs)),
			TotalTxs:   uint64(len(txs)),
			ParentHash: common.EmptyHash,
			GasLimit:   1e19,
		},
		Data: &types.Data{
			Txs: txs,
		},
	}

	return block
}

var serTxAccountToUTXO string = "f904b580f852cf4852c16c3232f849808502540be400a06cb74a2e1cb5c80076f0f3241cb17ce4a1c48a5aede9d3d5c46eae24ec1fc501a0599fbe47523a149a1ecacf7f1666ee90acf23daf8e61e224f6214536bcea88f2ea5842699137a517e2a082090b2fc114929526393191840d06a56b593e9ad3b1dfe00c14387fab0292ab809400000000000000000000000000000000000000009454fb1c7d0f011dd63b08f85ed7b518ab82028100a08d2e746c2caa306a57ea3349ddd4e049933065f50c4a01314b9994d29143f0ea8402faf08080f84582ec8ea016eeb095fc1f37e3c8004ae2f89bb1d0d12a892c8ef38685ec42a2e7b6f2b25aa06b4bcf38a5defd312218ad5948cb4f54524f10f2d3155afc56441797c216019ff9039af8d280a00000000000000000000000000000000000000000000000000000000000000000c0c0f865f863a00000000000000000000000000000000000000000000000000000000000000000a012e4c14bcbcc79ef000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000f844f842a00000000000000000000000000000000000000000000000000000000000000000a0599fbe47523a149a1ecacf7f1666ee90acf23daf8e61e224f6214536bcea88f280f902c3c0f902bdf902bac0a0e181ef4ae2511ea6db183eeecdf1d5510d224ca24331bd51e53da0889ac7cec7a0fc780fc612127862679cf3c15438db9253e992ce9670f05c0a028b0aa5a824d8a02c10fb57e73c19fbeccfb221d14abcb3e6499d4d48f53d35a56948c57e1a3b33a080144cc4b35cf42c7d1cba7204fe52c3f4cbea36274dcaf0540f67d2a82b0062a0482d25dbe62ef856469579255f2bf4b39342155c5ce0d8cf19134cfcf3981e09a070a32d738eb7bad948d65c2263366afe6855906ff6bc66cdd6f5a38fd3c4c008f8c6a0900a60686de972bfce2ba8b83549a1926ee8119f05e5b569d7dbbf3026f8fa3ba096b35acde266555b1198af0665f9deffb50301840eead9d3d764eab42068b0f9a07515d1af6f1a7ac4873371ecabc698238cf20d372800853d90bede6c36a13da0a0860ec7842ac816113918f4079cc8ba85c6f77f0b59b378a7d7c6aeec4135722ca0c4edad0c7943aa6dc414a59cf9bf7ffb206524419cb90b3f0ade70bcc3c23bc5a09aa20f613594ea6be45756c015f116d4b7bc0b1368bfdfc57c96f4810c8854c7f8c6a06a3e6ac018fae6efba95ab0e3d0a8d4ace65faf3c42bbd7c3fdad4d6ec35f1e6a0373a4c2f47f2fc8ad338941a2d30dfccefde1834aec9c8cb5b9b1ae24e0bc119a03814145be55ad6ee20ef23c69bc91ce0dbe95218261c4040e8c3a2bef3266398a01479172a4f6b8f2876c0ed36be225c11e4ac5ef131acc85b84cf7024633592fda0ebce563708c22d4f06062166acd2d12a5e6cf7c13da5473f2c096f02098b8271a0a92a8e52e90be9cf18e9e5d83244f1c0d48099b198d9b46c2ac9094b97ad3d2ea05c78b57e48772a917c7eba996d10e5f2acff0150ad4056c5779047c12869b807a07bacfbc8c9ab1fbf9c2586d3e318b3a512399b2023e08f58ea9b4d489450b60ea0782aec3f31e1d8269130f1f71cf89294bdfceb6c7cc2dd6f0218112219c4130ac0c0"
var serTxAccountToMix string = "f904f280f84dcf4852c16c3232f8448080a03745b1b5386568c5c9b224750641bbf64ea451dc9417dea914b8df4ff248330ca007880beb05ff90bbd335420174b6f07c47c2628985fe4ca6e8da7959dedbe762f86bc77cca059e4aa7f838947b6837189a3464d3c696069b2b42a9ae8e17dda18080a085598d81489afb5283fddcfb9cddb6a81b85ae91cbec9f75fb036238f50725895842699137a517e2a0742d3c403cdd70c25a8555f201a8fa1fe977ef96873f7817f4efda0e91cdd7fa809400000000000000000000000000000000000000009454fb1c7d0f011dd63b08f85ed7b518ab82028100a0cd970cc7b1f205cf46f0a569415a074dda48dc0f77a841a33ac600c3f547046b840773594080f84582ec8ea0152b35d3621a3ea7e4ecf579cd8b4398fc23b60629d942e287dd8266d654c7d3a022c60a440fa92e3fcc1dfd3cd1ab877326bfe23ffdc1d5973ff9259185691af4f9039af8d280a00000000000000000000000000000000000000000000000000000000000000000c0c0f865f863a00000000000000000000000000000000000000000000000000000000000000000a05ffb3600ddf3d01a000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000f844f842a00000000000000000000000000000000000000000000000000000000000000000a062f3347eb0218a1f5183a00fce1a7cdbbec5f3fe81be877315950baf013ee05f80f902c3c0f902bdf902bac0a0cc292034fca200f0a3b3aaaa63a52a7bd8ced893ee63166663125e7a882b23d9a03d81eef71a3a4c95906291b074fcd5e1196e40808d9b509a71177a4390a477a2a0a80172c15aa71dfb08ac5067661992f5f926d4b55142be82ed3d50fe7dffbec7a0dc7eec0dc68b532574acfabab1db5d8d1a034cb0e698a547bfbb74c29b182684a060f8df28bda1beeb1d55fc79117c8fdb0ef6f090561a129e6a1031a44d0bdf0ca0c39efbe4484cbe50a156dae824d01ccb6fe0e049b6d3803c03f25394d453990df8c6a0f20013639c922f13f7080606cfa8d4adf001092fff78c4b6ff0261588028efc1a0ae366d7434cc6eef59e3fab99dacbe6efbf09b03e372b208cc9af024f095b4d0a02fd683703d04d44705a791cede2a3d067aa3daf18e2f840f84a38a95da1ce6fda00b55158cc7c19dc0e4999c28c0eb39eb16a4eb88a85d8dc76742eb1d8a12afdaa02de8c8508784b6df1b40e34e7f2eef1211946e3ce84dbee0ae6f44b57fb43c78a09d50771db9468dba6ebdd0cf639086a75900bf893604cd5bb28c9e5b835e19c5f8c6a03c842fd3becbf562ca9bc00ecc715e7eb50b1859ea6386741208a1104c1a5624a073fa56bd57"

func genBlockAccountToMix(height uint64, statedb *state.StateDB) *types.Block {
	serTxAccountToMixTransfer := "f9065e80f854cf4852c16c3232f84b808701c90565ffe800a07a3d100a86f9bd1bda36bf974efb5f0ae33ca9da38d70a43d568e1547bb03306a0b55faeb1418d1a887475f20041cd8a8d184b799b03bd550da8875312b2ba20aff8e2c77cca059e4aa7f83e947b6837189a3464d3c696069b2b42a9ae8e17dda1865af3107a400080a0e70ea715c1e0cfd7145f496dd20ea10e991d678201c7d0f2df4944b3685b13e1c77cca059e4aa7f83e9408085a83232c4a3c2f9065f5bc1d93845fe8a4b586886c98b7600080a0d33d44d8ea542372c86c7f3397a7a9f8edb325f7d8cda12aec58b986baade4915842699137a517e2a0abe1734267929b878335827b1fcc751050baa7cc85feeddcdff7dd49cf2132d5805842699137a517e2a0da6774080cd4c4b4ba6224b95695d7011cdc4879503c268b8e327e951ef5654c809400000000000000000000000000000000000000009454fb1c7d0f011dd63b08f85ed7b518ab82028100a017217ff459b7d3f16b68bf987c7c7226e69a81aeadb3631980660334ffaca57f860246139ca80080f84582ec8da03d65e2427fac8bc2c3bb69990dbdcab1684de969bbabd735361737785a9f6ceca01848204165004a1d81e03fb49c693e245db0c31b4deeabd236353198f0118693f90486f9017b80a00000000000000000000000000000000000000000000000000000000000000000c0c0f8caf863a00000000000000000000000000000000000000000000000000000000000000000a05afd76f3b849b6e1000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000f863a00000000000000000000000000000000000000000000000000000000000000000a096c27196aac98d5a000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000f888f842a00000000000000000000000000000000000000000000000000000000000000000a0195fdd0108e4bfd66d0fcc21e8c1d2c2e7435f6bba58ff376ee741158f3f77bbf842a00000000000000000000000000000000000000000000000000000000000000000a06e388b862f4580bb976eb81497231cef8ba5e378e5e8321cf2465b4b3b4ffe2a80f90305c0f902fff902fcc0a0b399b897b7376f0beea319f0b22581d3fe88e0b250ed8e55c684198b754304e9a0a86e1e57fc5de6b0745aa533a90f70fd4c0788c24f88fb751206ea23596235f4a07409e4bf35b871c931542664140821743f92c4393beaa20487852f2d67d81ac4a0f1d73aef1f4216bdc24bb6a185a9b83c6ad0645452ed2f7d267c25a2add507e5a0ee9faa0dc5fbef1fa9e94b2fd21ae910877e14b8869a1383264110cda56e740ca085c7009f262cf367fe091b0a73f9051d9018d29b4924f5b81fb65dc89f66b400f8e7a0da0316f198b19b927699e09dddc6dea20f6efaf1e48a50b814527901d1a4c409a014742be152a4d53405411d0eff7150ce279c72d41fc881a1995f044106a44f99a064ef9a300476293cac8f263a79ea50e3a4307dfe996626a203a342b0acd9b818a04ba93ef6d5c99325934de8cb567c51e24e31d6aedb28d796f26fb1d3c8acbf03a07b083680ccaf346c8a6cc92ad7eea26e4bfe53d569f01e4d58c5ee05f1367a92a02abd09863923e259cd437daf10069c45cce15038e103cdb8cb4e9a431304001da00b62c66e4bad1541bbaf43843dbd37c1ba69aadedc4274099ec279856cf02714f8e7a00c224e0c06a8c6880bc197304d126b0ef0fff97e485c9bf537c92a797439f2cca052ebfb437b60295d7e6c3d580e3c139b2218b58152b9935b3a199d81e75b5e39a02c1ef05d6982a2a64402edbc14508fb8d3c6073e0ba596c9853322ea47f5b3c3a0ddb2cee14c277c6c584af11d954bac7ebdcced29b5f2ef3093371e58c9c2e40ea08e2aa37c1e013f8770706f2c000b56ab726e6b769266a3f78f2bdd3ebfb1918da0a3d10fe1087f27f4790165ce9606adba2ed74d2bcea3abe6a45bb9584e517134a066ed1a45cffd6a9c990e1d12f71b12fd45945f03c35a51924688ae3f70818933a071d3871b301dc1bf08afeed990a2b151e0b771f85ada4f5ec8cd9a26522bd90fa01cdf07649bdbd0165aeb0ac5a560ec7f88bd4a997acaee2458e944a0103d4c09a08ddcb56efbf1ed1fe92bda206996be649d5db7eac78ceaa4fc067b409ec63a02c0c0"
	// From:0x54fb1c7d0f011dd63b08f85ed7b518ab82028100 Amount 502500000000000
	//To:0x7b6837189a3464d3c696069b2b42a9ae8e17dda1,   Amount 100000000000000
	//To:0x08085a83232c4a3c2f9065f5bc1d93845fe8a4b5,   Amount 150000000000000
	//                                               Fee       :2500000000000

	txs := make(types.Txs, 0)
	utxoTx := genUTXOTransaction(serTxAccountToMixTransfer)

	txs = append(txs, utxoTx)
	if err := utxoTx.CheckBasic(nil); err != nil {
		log.Error("CheckBasic", "err", err)
	}

	block := &types.Block{
		Header: &types.Header{
			Height:     height,
			Coinbase:   coinbase,
			Time:       uint64(time.Now().Unix()),
			NumTxs:     uint64(len(txs)),
			TotalTxs:   uint64(len(txs)),
			ParentHash: common.EmptyHash,
			GasLimit:   1e19,
		},
		Data: &types.Data{
			Txs: txs,
		},
	}

	return block
}

func genBlockAccount2Account(height uint64, statedb *state.StateDB) *types.Block {
	//hash: 0xa6c0a339b527210528c0c56ec457c01ba97bd87ca33fd396a243dd48d03f15ba
	//{"from":"0x54fb1c7d0f011dd63b08f85ed7b518ab82028100","nonce":"0x0","dests":[{"addr":"0xa73810e519e1075010678d706533486d8ecc8000","amount":"0x56bc75e2d63100000","data":""}]}
	//serTxAccount2Account := "f90139f856cf4852c16c3232f84d80890572b7b98736c20000a00000000000000000000000000000000000000000000000000000000000000000a00449c9d0102bfb7232c1655f4833e60f1aef38ff2b9a535678debea611f2df97f84ac77cca059e4aa7f84194a73810e519e1075010678d706533486d8ecc800089056bc75e2d6310000080a04b1651c5abb75479ab0719edc3a221be8607e32ebf1b3f45a0b4db3348ca625f940000000000000000000000000000000000000000a02e16c7e81ec6dbd0577537006734f567063023ae61b265510fc8a0a449c8a779c08806f05b59d3b2000080f84582e3e6a0054ce9bb1f7944fc3af519780f28c9e5240ab2f2555570ccb937a4900c1a0b36a032577ac35f69ce2216119ee995080e46f9e510f2ce2c5bc0475791de7d6f88acccc580c0c0c080c5c0c0c0c0c0"
	txs := make(types.Txs, 0)
	//utxoTx := genUTXOTransaction(serTxAccount2Account)
    nonce := uint64(0)
    toAddr := common.HexToAddress("0xa73810e519e1075010678d706533486d8ecc8000")
    utxoTx := getUTXOTokenTx(accounts[0].PrivateKey, toAddr, common.EmptyAddress, nonce, big.NewInt(0).Mul(big.NewInt(1e9), big.NewInt(1e11)), nil)
	txs = append(txs, utxoTx)
	app, err := initApp()
	if err != nil {
		log.Error("initApp", "err", err)
		return nil
	}
	if err := utxoTx.CheckBasic(app); err != nil {
		log.Error("CheckBasic", "err", err)
		//return nil
	}

	block := &types.Block{
		Header: &types.Header{
			Height:     height,
			Coinbase:   coinbase,
			Time:       uint64(time.Now().Unix()),
			NumTxs:     uint64(len(txs)),
			TotalTxs:   uint64(len(txs)),
			ParentHash: common.EmptyHash,
			GasLimit:   1e19,
		},
		Data: &types.Data{
			Txs: txs,
		},
	}

	return block
}

func genBlockUTXOCallContract(height uint64, statedb *state.StateDB) *types.Block {
	txs := make(types.Txs, 0)
	nonce := uint64(0)
	demoTokenBIN := "60806040526012600160006101000a81548160ff021916908360ff16021790555034801561002c57600080fd5b50e4801561003957600080fd5b50600160009054906101000a900460ff1660ff16600a0a61271002600081905550600054e07fd1398bee19313d6bf672ccb116e51f4a1a947e91c757907f51fbb5b5e56c698f60003030600054604051808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200194505050505060405180910390a16107eb806101436000396000f300608060405260043610610083576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806306fdde0314610085578063313ce567146101225780633eaaf86b146101605780635d0268e61461019857806370a08231146101b8578063a4556fce1461021c578063d0ca623414610226575b005b34801561009157600080fd5b50e4801561009e57600080fd5b506100a7610230565b6040518080602001828103825283818151815260200191508051906020019080838360005b838110156100e75780820151818401526020810190506100cc565b50505050905090810190601f1680156101145780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561012e57600080fd5b50e4801561013b57600080fd5b5061014461026d565b604051808260ff1660ff16815260200191505060405180910390f35b34801561016c57600080fd5b50e4801561017957600080fd5b50610182610280565b6040518082815260200191505060405180910390f35b6101b660048036038101908080359060200190929190505050610286565b005b3480156101c457600080fd5b50e480156101d157600080fd5b50610206600480360381019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506103b0565b6040518082815260200191505060405180910390f35b6102246103e8565b005b61022e6105cf565b005b60606040805190810160405280600981526020017f44656d6f546f6b656e0000000000000000000000000000000000000000000000815250905090565b600160009054906101000a900460ff1681565b60005481565b3073ffffffffffffffffffffffffffffffffffffffff16e273ffffffffffffffffffffffffffffffffffffffff161415156102c057600080fd5b80e41480156102cf5750600081115b15156102da57600080fd5b7fd1398bee19313d6bf672ccb116e51f4a1a947e91c757907f51fbb5b5e56c698f3330e2e4604051808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200194505050505060405180910390a150565b60008173ffffffffffffffffffffffffffffffffffffffff163073ffffffffffffffffffffffffffffffffffffffff16e19050919050565b6000e41115156103f757600080fd5b3373ffffffffffffffffffffffffffffffffffffffff163073ffffffffffffffffffffffffffffffffffffffff16e4e37fd1398bee19313d6bf672ccb116e51f4a1a947e91c757907f51fbb5b5e56c698f3330e2e4604051808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200194505050505060405180910390a17fd1398bee19313d6bf672ccb116e51f4a1a947e91c757907f51fbb5b5e56c698f303330e4604051808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200194505050505060405180910390a1565b600080341115156105df57600080fd5b6002340290503373ffffffffffffffffffffffffffffffffffffffff163073ffffffffffffffffffffffffffffffffffffffff1682e37fd1398bee19313d6bf672ccb116e51f4a1a947e91c757907f51fbb5b5e56c698f3330600034604051808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200194505050505060405180910390a17fd1398bee19313d6bf672ccb116e51f4a1a947e91c757907f51fbb5b5e56c698f30333084604051808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200194505050505060405180910390a1505600a165627a7a7230582035535fe5dabdc379daafa902bd1a8b41026cc07fe489cad1a962214b964671dd0029"
	contractCreateTx := genContractCreateTx(accounts[0].Address, gasLimit, nonce, demoTokenBIN)
	nonce++
	txs = append(txs, contractCreateTx)
	fromaddr, _ := contractCreateTx.From()
	contractAddr := crypto.CreateAddress(fromaddr, contractCreateTx.Nonce(), contractCreateTx.Data())
	log.Info("genBlockUTXOCallContract", "contractAddr", contractAddr)

	utxoCallContract := "f905aeeb10c698b1d37308e3c180a03b0947117b33cebd8fc22e26cb203d95f0ca5c3b11d0b6843b387a7fa5290829f892c77cca059e4aa7f83d9437c9b94a0f4816ff9e209ff9fe56e2a094deefd785012a05f20080a0ab039980de18235e5fd3e89fb5999fbb7f6233e73981ef146d0472930b7068315842699137a517f843a005590df2358ba38889df848a79b07dd288246bb023356955e82ced56e67ef74980a01ccf4a4079c4ec97e51be09a9230475d326d5268887e99b965c1253ac8ba02ed940000000000000000000000000000000000000000a0cf95284e03a673acc0c99c5ef41bd519eafb19d50742b892af1b4a060b63a5f9e1a0070d50295717399610d976259cc878356412ee51b84f8a51692ac98d7aa150ee87b91c4fdb60800080f84582e3e5a0f2226baa77cb1fe182de8868ec8b7dc80269b9ba9083513b83479287a9f55c9da061dd4ba9549b5ca8dd18862be8009fcf8537eac91851fc9e6d02f382fc8688cef90443f8b003c0f865f863a0d21d19150211ef30e296fa502fda52c643c38b6cbf4cef03d09f6c7b6ca87c0fa0517e2473aa2c06e40f5ad7af5cf6864ae96b9ab74a978fdff2e8139b9c1de004a00000000000000000000000000000000000000000000000000000000000000000f844f842a00000000000000000000000000000000000000000000000000000000000000000a019574676cedbe237faed66ecaafabf3382bb9db163bbd86ce0408f8cf773f71180f9038ec0f902fef902fba067e5985fa37c9f0f361a11b2b174182f785528d51b6546b7141d0d0db271985ba0751c95495c3821e246df9ed4bfc43f5f3dd46b54a7ef98d6050b1967030a5735a05a8d9a76d6a93a697b243b2c9f66b14294dfc7afccc854297f7091ca64f2a969a01aca5c48a33a8127ae58b4491c4730653bee98df4d5ce84a9d02ab55486b4b74a037ca8238236b5998769a19170a9097f4fdf0bdc7fe135df18f3007848908c60fa0eeaf35c3e5b2fda5b93dd7eb5bec9a4ed0292938f986baaf00d37cce969e8407f8e7a0c8868fc4570c9e4866da08db338f9d6f0bac3850e1acf867144e62e7dfd5f5eda0387f00627937a6d67d2633eb2626cbcac79e3a6898a74994e706d72afc6a6ae8a063bd824073f905817d27cb9935330806feb0e4d962ef903466f540e52c5ede4ca06d2c5d83c52dfcb2db03136a01e612115ce9948c255b2677ed7e3e0d47f2ebfaa033666469052b3ceb6368a2be0661e5a05240ff1f5107a078c4565fb7d7ef2810a0ed76ae50b191d4374b28b019a034fd8ae21ff7f2553e4d053457acec68d21879a016e650415a5b7272f1408f7847150c8a69961132392340a5eddda91facf9fcb5f8e7a0ce4b67624309d62e704401e6b6cc27dfe2ce5d0b9ae8802a83aaee8bee0bb773a02e66f8141074ed6429aebffced2b22e571e96c40e0d4860a87cf6f52715b14d9a0cb4aff8ec9249014dadb077ac1988aff888789e79bf55f37da20b58c0dddd7bda0b3ca1aeb0f366248a967e853a023671d5af7bde28246d06ac5c3d3e0e13f6e04a0b5074a34f8b046e64aa19d1374904aa13a4c02dfdb5e335185e63c0a898fea85a087edc7dfaecf5fed3b8d31609de81e2c18c6fdc2ac9f6e597a72ed0642b8709ba0ec1399c9c00a45157d000c209a9a1a1f8dd386f399b3142b311b4334153c3dcda0d1e6fe08ed8d80afcce4101515d9f72e31724c53ebc95506ad62b06e8a77ef08a0e222fa5ca652d91fb75865f3e405e429145b96b335f6a594dc6d359c7afefa09a03821bda8d286828f59eff85c16999d1529ca4528d995246b2899b05847a02a0ee3e2c0a00000000000000000000000000000000000000000000000000000000000000000e1a0ef8d7abad7a9c6f97027aa570da72a8d1f4b364fa9d5600c5db8cd9bb3002da1f844f842a07b6b27b1ecd0f8c77003c87e9e90a029db188de403218de0c0a87be6e6ff1005a05d99aec15b0a07a5e0cb35c28ed79926db17a1e486bc9e4e69fafee38d97ad08"
	utxoTx := genUTXOTransaction(utxoCallContract)
	/*
		app, err := initApp()
		if err != nil {
			log.Error("initApp", "err", err)
			return nil
		}
		if err := utxoTx.CheckBasic(app); err != nil {
			log.Error("CheckBasic", "err", err)
		}
	*/
	txs = append(txs, utxoTx)

	block := &types.Block{
		Header: &types.Header{
			Height:     height,
			Coinbase:   coinbase,
			Time:       uint64(time.Now().Unix()),
			NumTxs:     uint64(len(txs)),
			TotalTxs:   uint64(len(txs)),
			ParentHash: common.EmptyHash,
			GasLimit:   1e19,
		},
		Data: &types.Data{
			Txs: txs,
		},
	}
	return block
}

//accounts[0]->subaddr[0]  =>  accounts[0]->subaddr[1]
func genBlockUTXO2UTXO(height uint64, statedb *state.StateDB, fee *big.Int) *types.Block {
	txs := make(types.Txs, 0)
    utxoInput := &types.UTXOInput{}
    utxoOutput := &types.UTXOOutput{}
    utxoTx := &types.UTXOTransaction{
        Inputs: []types.Input{utxoInput},
        Outputs: []types.Output{utxoOutput},
        Fee: fee, 
    }
    utxoTx.RCTSig.OutPk = make(lctypes.CtkeyV, 1)
    utxoTx.RCTSig.OutPk[0].Mask = lctypes.Key{}

	txs = append(txs, utxoTx)

	block := &types.Block{
		Header: &types.Header{
			Height:     height,
			Coinbase:   coinbase,
			Time:       uint64(time.Now().Unix()),
			NumTxs:     uint64(len(txs)),
			TotalTxs:   uint64(len(txs)),
			ParentHash: common.EmptyHash,
			GasLimit:   1e19,
		},
		Data: &types.Data{
			Txs: txs,
		},
	}
	return block
}

func genBlockAccountToContract(height uint64, statedb *state.StateDB) *types.Block {
	fmt.Println("account[0].address", accounts[0].Address.String())

	txs := make(types.Txs, 0)
	nonce := uint64(0)
	demoTokenBIN := "60806040526012600160006101000a81548160ff021916908360ff16021790555034801561002c57600080fd5b50e4801561003957600080fd5b50600160009054906101000a900460ff1660ff16600a0a61271002600081905550600054e07fd1398bee19313d6bf672ccb116e51f4a1a947e91c757907f51fbb5b5e56c698f60003030600054604051808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200194505050505060405180910390a16107eb806101436000396000f300608060405260043610610083576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806306fdde0314610085578063313ce567146101225780633eaaf86b146101605780635d0268e61461019857806370a08231146101b8578063a4556fce1461021c578063d0ca623414610226575b005b34801561009157600080fd5b50e4801561009e57600080fd5b506100a7610230565b6040518080602001828103825283818151815260200191508051906020019080838360005b838110156100e75780820151818401526020810190506100cc565b50505050905090810190601f1680156101145780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561012e57600080fd5b50e4801561013b57600080fd5b5061014461026d565b604051808260ff1660ff16815260200191505060405180910390f35b34801561016c57600080fd5b50e4801561017957600080fd5b50610182610280565b6040518082815260200191505060405180910390f35b6101b660048036038101908080359060200190929190505050610286565b005b3480156101c457600080fd5b50e480156101d157600080fd5b50610206600480360381019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506103b0565b6040518082815260200191505060405180910390f35b6102246103e8565b005b61022e6105cf565b005b60606040805190810160405280600981526020017f44656d6f546f6b656e0000000000000000000000000000000000000000000000815250905090565b600160009054906101000a900460ff1681565b60005481565b3073ffffffffffffffffffffffffffffffffffffffff16e273ffffffffffffffffffffffffffffffffffffffff161415156102c057600080fd5b80e41480156102cf5750600081115b15156102da57600080fd5b7fd1398bee19313d6bf672ccb116e51f4a1a947e91c757907f51fbb5b5e56c698f3330e2e4604051808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200194505050505060405180910390a150565b60008173ffffffffffffffffffffffffffffffffffffffff163073ffffffffffffffffffffffffffffffffffffffff16e19050919050565b6000e41115156103f757600080fd5b3373ffffffffffffffffffffffffffffffffffffffff163073ffffffffffffffffffffffffffffffffffffffff16e4e37fd1398bee19313d6bf672ccb116e51f4a1a947e91c757907f51fbb5b5e56c698f3330e2e4604051808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200194505050505060405180910390a17fd1398bee19313d6bf672ccb116e51f4a1a947e91c757907f51fbb5b5e56c698f303330e4604051808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200194505050505060405180910390a1565b600080341115156105df57600080fd5b6002340290503373ffffffffffffffffffffffffffffffffffffffff163073ffffffffffffffffffffffffffffffffffffffff1682e37fd1398bee19313d6bf672ccb116e51f4a1a947e91c757907f51fbb5b5e56c698f3330600034604051808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200194505050505060405180910390a17fd1398bee19313d6bf672ccb116e51f4a1a947e91c757907f51fbb5b5e56c698f30333084604051808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200194505050505060405180910390a1505600a165627a7a7230582035535fe5dabdc379daafa902bd1a8b41026cc07fe489cad1a962214b964671dd0029"
	contractCreateTx := genContractCreateTx(accounts[0].Address, gasLimit, nonce, demoTokenBIN)
	nonce++
	txs = append(txs, contractCreateTx)

	serTxAccountToUTXO := "f9013df856cf4852c16c3232f84d01890572c43566d008b800a00000000000000000000000000000000000000000000000000000000000000000a04a46c58b9c5dd33e27a81a6ef3d9ffe74121c2f2afe402cd305d7dab5bc5ed2ff84ec77cca059e4aa7f8459437c9b94a0f4816ff9e209ff9fe56e2a094deefd789056bc75e2d6310000084d0ca6234a04b1651c5abb75479ab0719edc3a221be8607e32ebf1b3f45a0b4db3348ca625f940000000000000000000000000000000000000000a09072581a50602b5b0a067358f9b9a419143eb7efa5565e19bc84c550a7bdcbf8c08806fcd7396cf8b80080f84582e3e5a099fc99e071d7eccbbcbfbf27981fa1d3dacb7c86935ba889ca114502b7ef83b9a074b07e15087e901065de877844ec9495981ce40d3dc4314584bf0ce7c5b563b6ccc580c0c0c080c5c0c0c0c0c0"
	utxoTx := genUTXOTransaction(serTxAccountToUTXO)
	nonce++
	//var app *LinkApplication
	app, err := initApp()
	if err != nil {
		log.Error("initApp", "err", err)
		return nil
	}
	if err := utxoTx.CheckBasic(app); err != nil {
		log.Error("CheckBasic", "err", err)
	}
	txs = append(txs, utxoTx)

	fromaddr, _ := contractCreateTx.From()
	contractAddr := crypto.CreateAddress(fromaddr, contractCreateTx.Nonce(), contractCreateTx.Data())
	log.Debug("", "contractAddr", contractAddr.String())
	demoTokenABI := `[{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"pure","type":"function"},{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint8"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"_totalSupply","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"value","type":"uint256"}],"name":"addOrder","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":true,"inputs":[{"name":"owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"exchangebytoken","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":false,"inputs":[],"name":"exchangebylk","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"inputs":[],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"payable":true,"stateMutability":"payable","type":"fallback"},{"anonymous":false,"inputs":[{"indexed":false,"name":"from","type":"address"},{"indexed":false,"name":"to","type":"address"},{"indexed":false,"name":"token","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"}]`

	//var cabi abi.ABI
	cabi, err := abi.JSON(bytes.NewReader([]byte(demoTokenABI)))
	if err != nil {
		panic(err)
	}

	method := "balanceOf"
	var data []byte
	data, err = cabi.Pack(method, common.HexToAddress("0x54fb1c7d0f011dd63b08f85ed7b518ab82028100")) ////sha3("name()")
	log.Debug("", "method data", fmt.Sprintf("0x%x", data))
	if err != nil {
		panic(err)
	}

	method = "exchangebylk"
	data, err = cabi.Pack(method) ////sha3("name()")
	log.Debug("", "method data", fmt.Sprintf("0x%x", data))

	if err != nil {
		panic(err)
	}
	//var contractCallTx *types.Transaction
	_, err = genTx(accounts[0], nonce, &contractAddr, big.NewInt(0), data)
	if err != nil {
		fmt.Println("gentx err", err)
	}
	//txs = append(txs, contractCallTx)

	block := &types.Block{
		Header: &types.Header{
			Height:     height,
			Coinbase:   coinbase,
			Time:       uint64(time.Now().Unix()),
			NumTxs:     uint64(len(txs)),
			TotalTxs:   uint64(len(txs)),
			ParentHash: common.EmptyHash,
			GasLimit:   1e19,
		},
		Data: &types.Data{
			Txs: txs,
		},
	}

	return block
}

func genBlockWithContractCreate(height uint64, statedb *state.StateDB) *types.Block {
	txs := make(types.Txs, 0)
	gasLimit := uint64(0)
	nonce := statedb.GetNonce(accounts[0].Address)

	ctx := genContractCreateTx(accounts[0].Address, 1000000, nonce, "../test/token/sol/SimpleToken.bin")
	ctx.Amount = new(big.Int).SetUint64(1)
	txs = append(txs, ctx)
	txs = append(txs, genContractCreateTx(accounts[0].Address, gasLimit, nonce+1, "../test/token/sol/SimpleToken.bin"))

	wasmCTX := genContractCreateTx(accounts[0].Address, gasLimit, nonce+2, "../test/token/tcvm/TestToken.bin")
	txs = append(txs, wasmCTX)

	contractAddr := crypto.CreateAddress(wasmCTX.FromAddr, wasmCTX.Nonce(), wasmCTX.Data())
	cut := genContractUpgradeTx(wasmCTX.FromAddr, contractAddr, wasmCTX.Nonce()+1, "../test/token/tcvm/TestToken.bin")
	txs = append(txs, cut)

	block := &types.Block{
		Header: &types.Header{
			Height:     height,
			Coinbase:   coinbase,
			Time:       uint64(time.Now().Unix()),
			NumTxs:     uint64(len(txs)),
			TotalTxs:   uint64(len(txs)),
			ParentHash: common.EmptyHash,
			GasLimit:   1e19,
		},
		Data: &types.Data{
			Txs: txs,
		},
	}

	return block
}

func genBlockWithTransactionContract(height uint64, statedb *state.StateDB) *types.Block {
	txs := make(types.Txs, 0)
	//gasLimit := uint64(0)
	nonce := statedb.GetNonce(accounts[0].Address)
	txs = append(txs, genTxForCreateContract(accounts[0], 10000000, nonce, "../test/token/sol/SimpleToken.bin"))
	txs = append(txs, genTxForCreateContract(accounts[0], 10000000, nonce+1, "../test/token/sol/SimpleToken.bin"))
	block := &types.Block{
		Header: &types.Header{
			Height:     height,
			Coinbase:   coinbase,
			Time:       uint64(time.Now().Unix()),
			NumTxs:     uint64(len(txs)),
			TotalTxs:   uint64(len(txs)),
			ParentHash: common.EmptyHash,
			GasLimit:   1e19,
		},
		Data: &types.Data{
			Txs: txs,
		},
	}

	return block
}

func genBlockWithLocalTransaction(height uint64) *types.Block {
	var maxTxNum = 1
	txs := make(types.Txs, 0)

	var nonce = uint64(0)

	for i := 0; i < maxTxNum; i++ {
		//tokenAddress := &common.Address{0}
		to := testToAddr
		amount := big.NewInt(1e18)
		gasLimit = types.CalNewAmountGas(amount, types.EverLiankeFee)
		signedTx, err := genTx(accounts[0], nonce, &to, amount, nil)
		if err != nil {
			panic(err)
		}
		txs = append(txs, signedTx)
		nonce++
	}

	block := &types.Block{
		Header: &types.Header{
			Height:     height,
			Coinbase:   coinbase,
			Time:       uint64(time.Now().Unix()),
			NumTxs:     uint64(len(txs)),
			TotalTxs:   uint64(len(txs)),
			ParentHash: common.EmptyHash,
			GasLimit:   1e19,
		},
		Data: &types.Data{
			Txs: txs,
		},
	}

	return block
}

func getUTXOTokenTx(skey *ecdsa.PrivateKey, toAddr common.Address, tokenID common.Address, nonce uint64, amount *big.Int, data []byte) *types.UTXOTransaction {
	addr := crypto.PubkeyToAddress(skey.PublicKey)
	accountSource := &types.AccountSourceEntry{
		From:   addr,
		Nonce:  nonce,
		Amount: amount,
	}
	transferGas := types.CalNewAmountGas(amount, types.EverLiankeFee) + 600000  //600000 for call contract gas 100000 + contract value fee 500000
	transferFee := big.NewInt(0).Mul(big.NewInt(types.ParGasPrice), big.NewInt(0).SetUint64(transferGas))
	accountDest := &types.AccountDestEntry{
		To:     toAddr,
		Amount: big.NewInt(0).Sub(amount, transferFee),
		Data:   data,
	}
	if !common.IsLKC(tokenID) {
		accountDest.Amount = amount
	}
	dest := []types.DestEntry{accountDest}

	utxoTx, _, err := types.NewAinTransaction(accountSource, dest, tokenID, nil)
	if err != nil {
		log.Error("getUTXOTx", "NewAinTransaction err", err)
		return nil
	}
	if !common.IsLKC(tokenID) {
		utxoTx.Fee = transferFee
	}
	err = utxoTx.Sign(types.GlobalSTDSigner, skey)
	if err != nil {
		log.Error("getUTXOTx", "Sign err", err)
		return nil
	}

	return utxoTx
}

func genBlockUTXOTokenTransaction(height uint64, statedb *state.StateDB) *types.Block {
	fmt.Println("account[0].address", accounts[0].Address.String())

	txs := make(types.Txs, 0)
	nonce := uint64(0)
	demoTokenBIN := "60806040526012600160006101000a81548160ff021916908360ff16021790555034801561002c57600080fd5b50e4801561003957600080fd5b50600160009054906101000a900460ff1660ff16600a0a61271002600081905550600054e07fd1398bee19313d6bf672ccb116e51f4a1a947e91c757907f51fbb5b5e56c698f60003030600054604051808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200194505050505060405180910390a16107eb806101436000396000f300608060405260043610610083576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806306fdde0314610085578063313ce567146101225780633eaaf86b146101605780635d0268e61461019857806370a08231146101b8578063a4556fce1461021c578063d0ca623414610226575b005b34801561009157600080fd5b50e4801561009e57600080fd5b506100a7610230565b6040518080602001828103825283818151815260200191508051906020019080838360005b838110156100e75780820151818401526020810190506100cc565b50505050905090810190601f1680156101145780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561012e57600080fd5b50e4801561013b57600080fd5b5061014461026d565b604051808260ff1660ff16815260200191505060405180910390f35b34801561016c57600080fd5b50e4801561017957600080fd5b50610182610280565b6040518082815260200191505060405180910390f35b6101b660048036038101908080359060200190929190505050610286565b005b3480156101c457600080fd5b50e480156101d157600080fd5b50610206600480360381019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506103b0565b6040518082815260200191505060405180910390f35b6102246103e8565b005b61022e6105cf565b005b60606040805190810160405280600981526020017f44656d6f546f6b656e0000000000000000000000000000000000000000000000815250905090565b600160009054906101000a900460ff1681565b60005481565b3073ffffffffffffffffffffffffffffffffffffffff16e273ffffffffffffffffffffffffffffffffffffffff161415156102c057600080fd5b80e41480156102cf5750600081115b15156102da57600080fd5b7fd1398bee19313d6bf672ccb116e51f4a1a947e91c757907f51fbb5b5e56c698f3330e2e4604051808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200194505050505060405180910390a150565b60008173ffffffffffffffffffffffffffffffffffffffff163073ffffffffffffffffffffffffffffffffffffffff16e19050919050565b6000e41115156103f757600080fd5b3373ffffffffffffffffffffffffffffffffffffffff163073ffffffffffffffffffffffffffffffffffffffff16e4e37fd1398bee19313d6bf672ccb116e51f4a1a947e91c757907f51fbb5b5e56c698f3330e2e4604051808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200194505050505060405180910390a17fd1398bee19313d6bf672ccb116e51f4a1a947e91c757907f51fbb5b5e56c698f303330e4604051808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200194505050505060405180910390a1565b600080341115156105df57600080fd5b6002340290503373ffffffffffffffffffffffffffffffffffffffff163073ffffffffffffffffffffffffffffffffffffffff1682e37fd1398bee19313d6bf672ccb116e51f4a1a947e91c757907f51fbb5b5e56c698f3330600034604051808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200194505050505060405180910390a17fd1398bee19313d6bf672ccb116e51f4a1a947e91c757907f51fbb5b5e56c698f30333084604051808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200194505050505060405180910390a1505600a165627a7a7230582035535fe5dabdc379daafa902bd1a8b41026cc07fe489cad1a962214b964671dd0029"
	contractCreateTx := genContractCreateTx(accounts[0].Address, gasLimit, nonce, demoTokenBIN)
	nonce++
	txs = append(txs, contractCreateTx)

	app, err := initApp()
	if err != nil {
		log.Error("initApp", "err", err)
		return nil
	}

	demoTokenABI := `[{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"pure","type":"function"},{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint8"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"_totalSupply","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"value","type":"uint256"}],"name":"addOrder","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":true,"inputs":[{"name":"owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"exchangebytoken","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":false,"inputs":[],"name":"exchangebylk","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"inputs":[],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"payable":true,"stateMutability":"payable","type":"fallback"},{"anonymous":false,"inputs":[{"indexed":false,"name":"from","type":"address"},{"indexed":false,"name":"to","type":"address"},{"indexed":false,"name":"token","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"}]`

	//var cabi abi.ABI
	cabi, err := abi.JSON(bytes.NewReader([]byte(demoTokenABI)))
	if err != nil {
		panic(err)
	}

	fromaddr, _ := contractCreateTx.From()
	contractAddr := crypto.CreateAddress(fromaddr, contractCreateTx.Nonce(), contractCreateTx.Data())
	log.Debug("", "contractAddr", contractAddr.String())
	var data []byte
	method := "exchangebylk"
	data, err = cabi.Pack(method)
	log.Debug("", "method data", fmt.Sprintf("0x%x", data))
	if err != nil {
		panic(err)
	}
	utxoTx1 := getUTXOTokenTx(accounts[0].PrivateKey, contractAddr, common.EmptyAddress, nonce, big.NewInt(10000000e11), data)
	nonce++
	if err := utxoTx1.CheckBasic(app); err != nil {
		log.Error("CheckBasic", "err", err)
	}
	txs = append(txs, utxoTx1)

	method = "exchangebytoken"
	data, err = cabi.Pack(method)
	log.Debug("", "method data", fmt.Sprintf("0x%x", data))
	if err != nil {
		panic(err)
	}
	utxoTx2 := getUTXOTokenTx(accounts[0].PrivateKey, contractAddr, contractAddr, nonce, big.NewInt(10000000e11), data)
	nonce++
	if err := utxoTx2.CheckBasic(app); err != nil {
		log.Error("CheckBasic", "err", err)
	}
	txs = append(txs, utxoTx2)

	block := &types.Block{
		Header: &types.Header{
			Height:     height,
			Coinbase:   coinbase,
			Time:       uint64(time.Now().Unix()),
			NumTxs:     uint64(len(txs)),
			TotalTxs:   uint64(len(txs)),
			ParentHash: common.EmptyHash,
			GasLimit:   1e19,
		},
		Data: &types.Data{
			Txs: txs,
		},
	}

	return block
}
