// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

import (
	"encoding/json"

	"github.com/ethereum-optimism/optimism/op-bindings/solc"
)

const L1CrossDomainMessengerStorageLayoutJSON = "{\"storage\":[{\"astId\":1000,\"contract\":\"src/L1/L1CrossDomainMessenger.sol:L1CrossDomainMessenger\",\"label\":\"spacer_0_0_20\",\"offset\":0,\"slot\":\"0\",\"type\":\"t_address\"},{\"astId\":1001,\"contract\":\"src/L1/L1CrossDomainMessenger.sol:L1CrossDomainMessenger\",\"label\":\"_initialized\",\"offset\":20,\"slot\":\"0\",\"type\":\"t_uint8\"},{\"astId\":1002,\"contract\":\"src/L1/L1CrossDomainMessenger.sol:L1CrossDomainMessenger\",\"label\":\"_initializing\",\"offset\":21,\"slot\":\"0\",\"type\":\"t_bool\"},{\"astId\":1003,\"contract\":\"src/L1/L1CrossDomainMessenger.sol:L1CrossDomainMessenger\",\"label\":\"spacer_1_0_1600\",\"offset\":0,\"slot\":\"1\",\"type\":\"t_array(t_uint256)50_storage\"},{\"astId\":1004,\"contract\":\"src/L1/L1CrossDomainMessenger.sol:L1CrossDomainMessenger\",\"label\":\"spacer_51_0_20\",\"offset\":0,\"slot\":\"51\",\"type\":\"t_address\"},{\"astId\":1005,\"contract\":\"src/L1/L1CrossDomainMessenger.sol:L1CrossDomainMessenger\",\"label\":\"spacer_52_0_1568\",\"offset\":0,\"slot\":\"52\",\"type\":\"t_array(t_uint256)49_storage\"},{\"astId\":1006,\"contract\":\"src/L1/L1CrossDomainMessenger.sol:L1CrossDomainMessenger\",\"label\":\"spacer_101_0_1\",\"offset\":0,\"slot\":\"101\",\"type\":\"t_bool\"},{\"astId\":1007,\"contract\":\"src/L1/L1CrossDomainMessenger.sol:L1CrossDomainMessenger\",\"label\":\"spacer_102_0_1568\",\"offset\":0,\"slot\":\"102\",\"type\":\"t_array(t_uint256)49_storage\"},{\"astId\":1008,\"contract\":\"src/L1/L1CrossDomainMessenger.sol:L1CrossDomainMessenger\",\"label\":\"spacer_151_0_32\",\"offset\":0,\"slot\":\"151\",\"type\":\"t_uint256\"},{\"astId\":1009,\"contract\":\"src/L1/L1CrossDomainMessenger.sol:L1CrossDomainMessenger\",\"label\":\"spacer_152_0_1568\",\"offset\":0,\"slot\":\"152\",\"type\":\"t_array(t_uint256)49_storage\"},{\"astId\":1010,\"contract\":\"src/L1/L1CrossDomainMessenger.sol:L1CrossDomainMessenger\",\"label\":\"spacer_201_0_32\",\"offset\":0,\"slot\":\"201\",\"type\":\"t_mapping(t_bytes32,t_bool)\"},{\"astId\":1011,\"contract\":\"src/L1/L1CrossDomainMessenger.sol:L1CrossDomainMessenger\",\"label\":\"spacer_202_0_32\",\"offset\":0,\"slot\":\"202\",\"type\":\"t_mapping(t_bytes32,t_bool)\"},{\"astId\":1012,\"contract\":\"src/L1/L1CrossDomainMessenger.sol:L1CrossDomainMessenger\",\"label\":\"successfulMessages\",\"offset\":0,\"slot\":\"203\",\"type\":\"t_mapping(t_bytes32,t_bool)\"},{\"astId\":1013,\"contract\":\"src/L1/L1CrossDomainMessenger.sol:L1CrossDomainMessenger\",\"label\":\"xDomainMsgSender\",\"offset\":0,\"slot\":\"204\",\"type\":\"t_address\"},{\"astId\":1014,\"contract\":\"src/L1/L1CrossDomainMessenger.sol:L1CrossDomainMessenger\",\"label\":\"msgNonce\",\"offset\":0,\"slot\":\"205\",\"type\":\"t_uint240\"},{\"astId\":1015,\"contract\":\"src/L1/L1CrossDomainMessenger.sol:L1CrossDomainMessenger\",\"label\":\"failedMessages\",\"offset\":0,\"slot\":\"206\",\"type\":\"t_mapping(t_bytes32,t_bool)\"},{\"astId\":1016,\"contract\":\"src/L1/L1CrossDomainMessenger.sol:L1CrossDomainMessenger\",\"label\":\"__gap\",\"offset\":0,\"slot\":\"207\",\"type\":\"t_array(t_uint256)44_storage\"},{\"astId\":1017,\"contract\":\"src/L1/L1CrossDomainMessenger.sol:L1CrossDomainMessenger\",\"label\":\"superchainConfig\",\"offset\":0,\"slot\":\"251\",\"type\":\"t_contract(SuperchainConfig)1018\"}],\"types\":{\"t_address\":{\"encoding\":\"inplace\",\"label\":\"address\",\"numberOfBytes\":\"20\"},\"t_array(t_uint256)44_storage\":{\"encoding\":\"inplace\",\"label\":\"uint256[44]\",\"numberOfBytes\":\"1408\",\"base\":\"t_uint256\"},\"t_array(t_uint256)49_storage\":{\"encoding\":\"inplace\",\"label\":\"uint256[49]\",\"numberOfBytes\":\"1568\",\"base\":\"t_uint256\"},\"t_array(t_uint256)50_storage\":{\"encoding\":\"inplace\",\"label\":\"uint256[50]\",\"numberOfBytes\":\"1600\",\"base\":\"t_uint256\"},\"t_bool\":{\"encoding\":\"inplace\",\"label\":\"bool\",\"numberOfBytes\":\"1\"},\"t_bytes32\":{\"encoding\":\"inplace\",\"label\":\"bytes32\",\"numberOfBytes\":\"32\"},\"t_contract(SuperchainConfig)1018\":{\"encoding\":\"inplace\",\"label\":\"contract SuperchainConfig\",\"numberOfBytes\":\"20\"},\"t_mapping(t_bytes32,t_bool)\":{\"encoding\":\"mapping\",\"label\":\"mapping(bytes32 =\u003e bool)\",\"numberOfBytes\":\"32\",\"key\":\"t_bytes32\",\"value\":\"t_bool\"},\"t_uint240\":{\"encoding\":\"inplace\",\"label\":\"uint240\",\"numberOfBytes\":\"30\"},\"t_uint256\":{\"encoding\":\"inplace\",\"label\":\"uint256\",\"numberOfBytes\":\"32\"},\"t_uint8\":{\"encoding\":\"inplace\",\"label\":\"uint8\",\"numberOfBytes\":\"1\"}}}"

var L1CrossDomainMessengerStorageLayout = new(solc.StorageLayout)

var L1CrossDomainMessengerDeployedBin = "0x6080604052600436106101755760003560e01c80636425666b116100cb578063a4e7f8bd1161007f578063c4d66de811610059578063c4d66de814610454578063d764ad0b14610474578063ecc704281461048757600080fd5b8063a4e7f8bd146103d4578063b1b1b20914610404578063b28ade251461043457600080fd5b806383a74074116100b057806383a74074146103895780638cbeeef21461029a5780639fce812c146103a057600080fd5b80636425666b146103415780636e296e451461037457600080fd5b80633dbb202b1161012d57806354fd4d501161010757806354fd4d50146102b05780635644cfdf146103065780635c975abb1461031c57600080fd5b80633dbb202b1461025d5780633f827a5a146102725780634c1d6a691461029a57600080fd5b80630ff754ea1161015e5780630ff754ea146101c25780632828d7e81461021b57806335e80ab31461023057600080fd5b8063028f85f71461017a5780630c568498146101ad575b600080fd5b34801561018657600080fd5b5061018f601081565b60405167ffffffffffffffff90911681526020015b60405180910390f35b3480156101b957600080fd5b5061018f603f81565b3480156101ce57600080fd5b506101f67f000000000000000000000000000000000000000000000000000000000000000081565b60405173ffffffffffffffffffffffffffffffffffffffff90911681526020016101a4565b34801561022757600080fd5b5061018f604081565b34801561023c57600080fd5b5060fb546101f69073ffffffffffffffffffffffffffffffffffffffff1681565b61027061026b366004611ad8565b6104ec565b005b34801561027e57600080fd5b50610287600181565b60405161ffff90911681526020016101a4565b3480156102a657600080fd5b5061018f619c4081565b3480156102bc57600080fd5b506102f96040518060400160405280600581526020017f322e322e3000000000000000000000000000000000000000000000000000000081525081565b6040516101a49190611baa565b34801561031257600080fd5b5061018f61138881565b34801561032857600080fd5b50610331610750565b60405190151581526020016101a4565b34801561034d57600080fd5b507f00000000000000000000000000000000000000000000000000000000000000006101f6565b34801561038057600080fd5b506101f66107e9565b34801561039557600080fd5b5061018f62030d4081565b3480156103ac57600080fd5b506101f67f000000000000000000000000000000000000000000000000000000000000000081565b3480156103e057600080fd5b506103316103ef366004611bc4565b60ce6020526000908152604090205460ff1681565b34801561041057600080fd5b5061033161041f366004611bc4565b60cb6020526000908152604090205460ff1681565b34801561044057600080fd5b5061018f61044f366004611bdd565b6108d5565b34801561046057600080fd5b5061027061046f366004611c31565b610943565b610270610482366004611c4e565b610b81565b34801561049357600080fd5b506104de60cd547dffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff167e010000000000000000000000000000000000000000000000000000000000001790565b6040519081526020016101a4565b6106257f000000000000000000000000000000000000000000000000000000000000000061051b8585856108d5565b347fd764ad0b0000000000000000000000000000000000000000000000000000000061058760cd547dffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff167e010000000000000000000000000000000000000000000000000000000000001790565b338a34898c8c6040516024016105a39796959493929190611d1d565b604080517fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe08184030181529190526020810180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff167fffffffff0000000000000000000000000000000000000000000000000000000090931692909217909152611558565b8373ffffffffffffffffffffffffffffffffffffffff167fcb0f7ffd78f9aee47a248fae8db181db6eee833039123e026dcbff529522e52a3385856106aa60cd547dffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff167e010000000000000000000000000000000000000000000000000000000000001790565b866040516106bc959493929190611d7c565b60405180910390a260405134815233907f8ebb2ec2465bdb2a06a66fc37a0963af8a2a6a1479d81d56fdb8cbb98096d5469060200160405180910390a2505060cd80547dffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff808216600101167fffff0000000000000000000000000000000000000000000000000000000000009091161790555050565b60fb54604080517f5c975abb000000000000000000000000000000000000000000000000000000008152905160009273ffffffffffffffffffffffffffffffffffffffff1691635c975abb9160048083019260209291908290030181865afa1580156107c0573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906107e49190611dca565b905090565b60cc5460009073ffffffffffffffffffffffffffffffffffffffff167fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff2153016108b8576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603560248201527f43726f7373446f6d61696e4d657373656e6765723a2078446f6d61696e4d657360448201527f7361676553656e646572206973206e6f7420736574000000000000000000000060648201526084015b60405180910390fd5b5060cc5473ffffffffffffffffffffffffffffffffffffffff1690565b6000611388619c4080603f6108f1604063ffffffff8816611e1b565b6108fb9190611e4b565b610906601088611e1b565b6109139062030d40611e99565b61091d9190611e99565b6109279190611e99565b6109319190611e99565b61093b9190611e99565b949350505050565b6000547501000000000000000000000000000000000000000000900460ff161580801561098e575060005460017401000000000000000000000000000000000000000090910460ff16105b806109c05750303b1580156109c0575060005474010000000000000000000000000000000000000000900460ff166001145b610a4c576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602e60248201527f496e697469616c697a61626c653a20636f6e747261637420697320616c72656160448201527f647920696e697469616c697a656400000000000000000000000000000000000060648201526084016108af565b600080547fffffffffffffffffffffff00ffffffffffffffffffffffffffffffffffffffff16740100000000000000000000000000000000000000001790558015610ad257600080547fffffffffffffffffffff00ffffffffffffffffffffffffffffffffffffffffff1675010000000000000000000000000000000000000000001790555b60fb80547fffffffffffffffffffffffff00000000000000000000000000000000000000001673ffffffffffffffffffffffffffffffffffffffff8416179055610b1a61160d565b8015610b7d57600080547fffffffffffffffffffff00ffffffffffffffffffffffffffffffffffffffffff169055604051600181527f7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb38474024989060200160405180910390a15b5050565b610b89610750565b15610bf0576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601c60248201527f43726f7373446f6d61696e4d657373656e6765723a207061757365640000000060448201526064016108af565b60f087901c60028110610cab576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152604d60248201527f43726f7373446f6d61696e4d657373656e6765723a206f6e6c7920766572736960448201527f6f6e2030206f722031206d657373616765732061726520737570706f7274656460648201527f20617420746869732074696d6500000000000000000000000000000000000000608482015260a4016108af565b8061ffff16600003610da0576000610cfc878986868080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152508f9250611704915050565b600081815260cb602052604090205490915060ff1615610d9e576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603760248201527f43726f7373446f6d61696e4d657373656e6765723a206c65676163792077697460448201527f6864726177616c20616c72656164792072656c6179656400000000000000000060648201526084016108af565b505b6000610de6898989898989898080601f01602080910402602001604051908101604052809392919081815260200183838082843760009201919091525061172392505050565b9050610df0611746565b15610e2857853414610e0457610e04611ec5565b600081815260ce602052604090205460ff1615610e2357610e23611ec5565b610f7a565b3415610edc576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152605060248201527f43726f7373446f6d61696e4d657373656e6765723a2076616c7565206d75737460448201527f206265207a65726f20756e6c657373206d6573736167652069732066726f6d2060648201527f612073797374656d206164647265737300000000000000000000000000000000608482015260a4016108af565b600081815260ce602052604090205460ff16610f7a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603060248201527f43726f7373446f6d61696e4d657373656e6765723a206d65737361676520636160448201527f6e6e6f74206265207265706c617965640000000000000000000000000000000060648201526084016108af565b610f838761186a565b15611036576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152604360248201527f43726f7373446f6d61696e4d657373656e6765723a2063616e6e6f742073656e60448201527f64206d65737361676520746f20626c6f636b65642073797374656d206164647260648201527f6573730000000000000000000000000000000000000000000000000000000000608482015260a4016108af565b600081815260cb602052604090205460ff16156110d5576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603660248201527f43726f7373446f6d61696e4d657373656e6765723a206d65737361676520686160448201527f7320616c7265616479206265656e2072656c617965640000000000000000000060648201526084016108af565b6110f6856110e7611388619c40611e99565b67ffffffffffffffff166118e1565b158061111c575060cc5473ffffffffffffffffffffffffffffffffffffffff1661dead14155b1561123557600081815260ce602052604080822080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff001660011790555182917f99d0e048484baa1b1540b1367cb128acd7ab2946d1ed91ec10e3c85e4bf51b8f91a27fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff320161122e576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602d60248201527f43726f7373446f6d61696e4d657373656e6765723a206661696c656420746f2060448201527f72656c6179206d6573736167650000000000000000000000000000000000000060648201526084016108af565b5050611533565b60cc80547fffffffffffffffffffffffff00000000000000000000000000000000000000001673ffffffffffffffffffffffffffffffffffffffff8a1617905560006112c688619c405a6112899190611ef4565b8988888080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152506118ff92505050565b60cc80547fffffffffffffffffffffffff00000000000000000000000000000000000000001661dead1790559050801561142257600082815260cb602052604090205460ff16156113bf576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152604560248201527f43726f7373446f6d61696e4d657373656e6765723a206d65737361676520686160448201527f7320616c7265616479206265656e2072656c6179656420766961207265656e7460648201527f72616e6379000000000000000000000000000000000000000000000000000000608482015260a4016108af565b600082815260cb602052604080822080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff001660011790555183917f4641df4a962071e12719d8c8c8e5ac7fc4d97b927346a3d7a335b1f7517e133c91a261152f565b600082815260ce602052604080822080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff001660011790555183917f99d0e048484baa1b1540b1367cb128acd7ab2946d1ed91ec10e3c85e4bf51b8f91a27fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff320161152f576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602d60248201527f43726f7373446f6d61696e4d657373656e6765723a206661696c656420746f2060448201527f72656c6179206d6573736167650000000000000000000000000000000000000060648201526084016108af565b5050505b50505050505050565b73ffffffffffffffffffffffffffffffffffffffff163b151590565b6040517fe9e05c4200000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff7f0000000000000000000000000000000000000000000000000000000000000000169063e9e05c429084906115d5908890839089906000908990600401611f0b565b6000604051808303818588803b1580156115ee57600080fd5b505af1158015611602573d6000803e3d6000fd5b505050505050505050565b6000547501000000000000000000000000000000000000000000900460ff166116b8576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602b60248201527f496e697469616c697a61626c653a20636f6e7472616374206973206e6f74206960448201527f6e697469616c697a696e6700000000000000000000000000000000000000000060648201526084016108af565b60cc5473ffffffffffffffffffffffffffffffffffffffff166117025760cc80547fffffffffffffffffffffffff00000000000000000000000000000000000000001661dead1790555b565b600061171285858585611919565b805190602001209050949350505050565b60006117338787878787876119b2565b8051906020012090509695505050505050565b60003373ffffffffffffffffffffffffffffffffffffffff7f0000000000000000000000000000000000000000000000000000000000000000161480156107e457507f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff167f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff16639bf62d826040518163ffffffff1660e01b8152600401602060405180830381865afa15801561182a573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061184e9190611f63565b73ffffffffffffffffffffffffffffffffffffffff1614905090565b600073ffffffffffffffffffffffffffffffffffffffff82163014806118db57507f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff16145b92915050565b600080603f83619c4001026040850201603f5a021015949350505050565b600080600080845160208601878a8af19695505050505050565b6060848484846040516024016119329493929190611f80565b604080517fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe08184030181529190526020810180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff167fcbd4ece9000000000000000000000000000000000000000000000000000000001790529050949350505050565b60608686868686866040516024016119cf96959493929190611fca565b604080517fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe08184030181529190526020810180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff167fd764ad0b0000000000000000000000000000000000000000000000000000000017905290509695505050505050565b73ffffffffffffffffffffffffffffffffffffffff81168114611a7357600080fd5b50565b60008083601f840112611a8857600080fd5b50813567ffffffffffffffff811115611aa057600080fd5b602083019150836020828501011115611ab857600080fd5b9250929050565b803563ffffffff81168114611ad357600080fd5b919050565b60008060008060608587031215611aee57600080fd5b8435611af981611a51565b9350602085013567ffffffffffffffff811115611b1557600080fd5b611b2187828801611a76565b9094509250611b34905060408601611abf565b905092959194509250565b6000815180845260005b81811015611b6557602081850181015186830182015201611b49565b81811115611b77576000602083870101525b50601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0169290920160200192915050565b602081526000611bbd6020830184611b3f565b9392505050565b600060208284031215611bd657600080fd5b5035919050565b600080600060408486031215611bf257600080fd5b833567ffffffffffffffff811115611c0957600080fd5b611c1586828701611a76565b9094509250611c28905060208501611abf565b90509250925092565b600060208284031215611c4357600080fd5b8135611bbd81611a51565b600080600080600080600060c0888a031215611c6957600080fd5b873596506020880135611c7b81611a51565b95506040880135611c8b81611a51565b9450606088013593506080880135925060a088013567ffffffffffffffff811115611cb557600080fd5b611cc18a828b01611a76565b989b979a50959850939692959293505050565b8183528181602085013750600060208284010152600060207fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0601f840116840101905092915050565b878152600073ffffffffffffffffffffffffffffffffffffffff808916602084015280881660408401525085606083015263ffffffff8516608083015260c060a0830152611d6f60c083018486611cd4565b9998505050505050505050565b73ffffffffffffffffffffffffffffffffffffffff86168152608060208201526000611dac608083018688611cd4565b905083604083015263ffffffff831660608301529695505050505050565b600060208284031215611ddc57600080fd5b81518015158114611bbd57600080fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b600067ffffffffffffffff80831681851681830481118215151615611e4257611e42611dec565b02949350505050565b600067ffffffffffffffff80841680611e8d577f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b92169190910492915050565b600067ffffffffffffffff808316818516808303821115611ebc57611ebc611dec565b01949350505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052600160045260246000fd5b600082821015611f0657611f06611dec565b500390565b73ffffffffffffffffffffffffffffffffffffffff8616815284602082015267ffffffffffffffff84166040820152821515606082015260a060808201526000611f5860a0830184611b3f565b979650505050505050565b600060208284031215611f7557600080fd5b8151611bbd81611a51565b600073ffffffffffffffffffffffffffffffffffffffff808716835280861660208401525060806040830152611fb96080830185611b3f565b905082606083015295945050505050565b868152600073ffffffffffffffffffffffffffffffffffffffff808816602084015280871660408401525084606083015283608083015260c060a083015261201560c0830184611b3f565b9897505050505050505056fea164736f6c634300080f000a"


func init() {
	if err := json.Unmarshal([]byte(L1CrossDomainMessengerStorageLayoutJSON), L1CrossDomainMessengerStorageLayout); err != nil {
		panic(err)
	}

	layouts["L1CrossDomainMessenger"] = L1CrossDomainMessengerStorageLayout
	deployedBytecodes["L1CrossDomainMessenger"] = L1CrossDomainMessengerDeployedBin
	immutableReferences["L1CrossDomainMessenger"] = true
}
