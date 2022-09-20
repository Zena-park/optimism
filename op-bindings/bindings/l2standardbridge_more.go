// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

import (
	"encoding/json"

	"github.com/ethereum-optimism/optimism/op-bindings/solc"
)

const L2StandardBridgeStorageLayoutJSON = "{\"storage\":[{\"astId\":26270,\"contract\":\"contracts/L2/L2StandardBridge.sol:L2StandardBridge\",\"label\":\"spacer_0_0_32\",\"offset\":0,\"slot\":\"0\",\"type\":\"t_bytes32\"},{\"astId\":26273,\"contract\":\"contracts/L2/L2StandardBridge.sol:L2StandardBridge\",\"label\":\"spacer_1_0_32\",\"offset\":0,\"slot\":\"1\",\"type\":\"t_bytes32\"},{\"astId\":26280,\"contract\":\"contracts/L2/L2StandardBridge.sol:L2StandardBridge\",\"label\":\"deposits\",\"offset\":0,\"slot\":\"2\",\"type\":\"t_mapping(t_address,t_mapping(t_address,t_uint256))\"},{\"astId\":26285,\"contract\":\"contracts/L2/L2StandardBridge.sol:L2StandardBridge\",\"label\":\"__gap\",\"offset\":0,\"slot\":\"3\",\"type\":\"t_array(t_uint256)47_storage\"}],\"types\":{\"t_address\":{\"encoding\":\"inplace\",\"label\":\"address\",\"numberOfBytes\":\"20\"},\"t_array(t_uint256)47_storage\":{\"encoding\":\"inplace\",\"label\":\"uint256[47]\",\"numberOfBytes\":\"1504\"},\"t_bytes32\":{\"encoding\":\"inplace\",\"label\":\"bytes32\",\"numberOfBytes\":\"32\"},\"t_mapping(t_address,t_mapping(t_address,t_uint256))\":{\"encoding\":\"mapping\",\"label\":\"mapping(address =\u003e mapping(address =\u003e uint256))\",\"numberOfBytes\":\"32\",\"key\":\"t_address\",\"value\":\"t_mapping(t_address,t_uint256)\"},\"t_mapping(t_address,t_uint256)\":{\"encoding\":\"mapping\",\"label\":\"mapping(address =\u003e uint256)\",\"numberOfBytes\":\"32\",\"key\":\"t_address\",\"value\":\"t_uint256\"},\"t_uint256\":{\"encoding\":\"inplace\",\"label\":\"uint256\",\"numberOfBytes\":\"32\"}}}"

var L2StandardBridgeStorageLayout = new(solc.StorageLayout)

var L2StandardBridgeDeployedBin = "0x6080604052600436106100e15760003560e01c8063662a633a1161007f578063a3a7954811610059578063a3a7954814610322578063af565a1314610335578063c89701a214610355578063e11013dd1461038957600080fd5b8063662a633a146102a957806387087623146102bc5780638f601f66146102dc57600080fd5b806332b7006d116100bb57806332b7006d146101fb5780633cb747bf1461020e578063540abf731461026757806354fd4d501461028757600080fd5b80630166a07a146101a057806309fc8843146101d55780631635f5fd146101e857600080fd5b3661019b57333b1561017a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603760248201527f5374616e646172644272696467653a2066756e6374696f6e2063616e206f6e6c60448201527f792062652063616c6c65642066726f6d20616e20454f4100000000000000000060648201526084015b60405180910390fd5b61019933333462030d406040518060200160405280600081525061039c565b005b600080fd5b3480156101ac57600080fd5b506101c06101bb366004612365565b61054c565b60405190151581526020015b60405180910390f35b6101996101e3366004612416565b6108dd565b6101996101f6366004612469565b6109b4565b6101996102093660046124dc565b610dd5565b34801561021a57600080fd5b506102427f000000000000000000000000000000000000000000000000000000000000000081565b60405173ffffffffffffffffffffffffffffffffffffffff90911681526020016101cc565b34801561027357600080fd5b50610199610282366004612530565b610e7a565b34801561029357600080fd5b5061029c610e93565b6040516101cc919061261d565b6101996102b7366004612365565b610f36565b3480156102c857600080fd5b506101996102d7366004612630565b611149565b3480156102e857600080fd5b506103146102f73660046126b3565b600260209081526000928352604080842090915290825290205481565b6040519081526020016101cc565b610199610330366004612630565b6111e8565b34801561034157600080fd5b506101996103503660046126ec565b6111f7565b34801561036157600080fd5b506102427f000000000000000000000000000000000000000000000000000000000000000081565b61019961039736600461273d565b61150a565b8373ffffffffffffffffffffffffffffffffffffffff168573ffffffffffffffffffffffffffffffffffffffff167f2849b43074093a05396b6f2a937dee8565b15a48a7b3d4bffb732a5017380af585846040516103fb9291906127a0565b60405180910390a37f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff16633dbb202b847f0000000000000000000000000000000000000000000000000000000000000000631635f5fd60e01b8989898860405160240161048094939291906127b9565b604080517fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe08184030181529181526020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff167fffffffff000000000000000000000000000000000000000000000000000000009485161790525160e086901b909216825261051392918890600401612802565b6000604051808303818588803b15801561052c57600080fd5b505af1158015610540573d6000803e3d6000fd5b50505050505050505050565b60003373ffffffffffffffffffffffffffffffffffffffff7f00000000000000000000000000000000000000000000000000000000000000001614801561066c57507f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff167f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff16636e296e456040518163ffffffff1660e01b8152600401602060405180830381865afa158015610630573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906106549190612847565b73ffffffffffffffffffffffffffffffffffffffff16145b61071e576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152604160248201527f5374616e646172644272696467653a2066756e6374696f6e2063616e206f6e6c60448201527f792062652063616c6c65642066726f6d20746865206f7468657220627269646760648201527f6500000000000000000000000000000000000000000000000000000000000000608482015260a401610171565b6040517faf565a1300000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff808a16600483015280891660248301528616604482015260648101859052309063af565a1390608401600060405180830381600087803b15801561079c57600080fd5b505af19250505080156107ad575060015b61084c576107c288888789886000898961154d565b8573ffffffffffffffffffffffffffffffffffffffff168773ffffffffffffffffffffffffffffffffffffffff168973ffffffffffffffffffffffffffffffffffffffff167f2755817676249910615f0a6a240ad225abe5343df8d527f7294c4af36a92009a8888888860405161083c94939291906128ad565b60405180910390a45060006108d2565b8573ffffffffffffffffffffffffffffffffffffffff168773ffffffffffffffffffffffffffffffffffffffff168973ffffffffffffffffffffffffffffffffffffffff167fd59c65b35445225835c83f50b6ede06a7be047d22e357073e250d9af537518cd888888886040516108c694939291906128ad565b60405180910390a45060015b979650505050505050565b333b1561096c576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603760248201527f5374616e646172644272696467653a2066756e6374696f6e2063616e206f6e6c60448201527f792062652063616c6c65642066726f6d20616e20454f410000000000000000006064820152608401610171565b6109af3333348686868080601f01602080910402602001604051908101604052809392919081815260200183838082843760009201919091525061039c92505050565b505050565b3373ffffffffffffffffffffffffffffffffffffffff7f000000000000000000000000000000000000000000000000000000000000000016148015610ad257507f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff167f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff16636e296e456040518163ffffffff1660e01b8152600401602060405180830381865afa158015610a96573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610aba9190612847565b73ffffffffffffffffffffffffffffffffffffffff16145b610b84576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152604160248201527f5374616e646172644272696467653a2066756e6374696f6e2063616e206f6e6c60448201527f792062652063616c6c65642066726f6d20746865206f7468657220627269646760648201527f6500000000000000000000000000000000000000000000000000000000000000608482015260a401610171565b823414610c13576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603a60248201527f5374616e646172644272696467653a20616d6f756e742073656e7420646f657360448201527f206e6f74206d6174636820616d6f756e742072657175697265640000000000006064820152608401610171565b3073ffffffffffffffffffffffffffffffffffffffff851603610cb8576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602360248201527f5374616e646172644272696467653a2063616e6e6f742073656e6420746f207360448201527f656c6600000000000000000000000000000000000000000000000000000000006064820152608401610171565b8373ffffffffffffffffffffffffffffffffffffffff168573ffffffffffffffffffffffffffffffffffffffff167f31b2166ff604fc5672ea5df08a78081d2bc6d746cadce880747f3643d819e83d858585604051610d19939291906128e3565b60405180910390a36000610d3e855a866040518060200160405280600081525061170e565b905080610dcd576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602360248201527f5374616e646172644272696467653a20455448207472616e736665722066616960448201527f6c656400000000000000000000000000000000000000000000000000000000006064820152608401610171565b505050505050565b333b15610e64576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603760248201527f5374616e646172644272696467653a2066756e6374696f6e2063616e206f6e6c60448201527f792062652063616c6c65642066726f6d20616e20454f410000000000000000006064820152608401610171565b610e7385333387878787611728565b5050505050565b610e8a878733888888888861195f565b50505050505050565b6060610ebe7f0000000000000000000000000000000000000000000000000000000000000000611bff565b610ee77f0000000000000000000000000000000000000000000000000000000000000000611bff565b610f107f0000000000000000000000000000000000000000000000000000000000000000611bff565b604051602001610f2293929190612906565b604051602081830303815290604052905090565b73ffffffffffffffffffffffffffffffffffffffff8716158015610f83575073ffffffffffffffffffffffffffffffffffffffff861673deaddeaddeaddeaddeaddeaddeaddeaddead0000145b1561101c57610f9585858585856109b4565b8473ffffffffffffffffffffffffffffffffffffffff168673ffffffffffffffffffffffffffffffffffffffff168873ffffffffffffffffffffffffffffffffffffffff167fb0444523268717a02698be47d0803aa7468c00acbed2f8bd93a0459cde61dd898787878760405161100f94939291906128ad565b60405180910390a4610e8a565b600061102d8789888888888861054c565b905080156110bc578573ffffffffffffffffffffffffffffffffffffffff168773ffffffffffffffffffffffffffffffffffffffff168973ffffffffffffffffffffffffffffffffffffffff167fb0444523268717a02698be47d0803aa7468c00acbed2f8bd93a0459cde61dd89888888886040516110af94939291906128ad565b60405180910390a461113f565b8573ffffffffffffffffffffffffffffffffffffffff168773ffffffffffffffffffffffffffffffffffffffff168973ffffffffffffffffffffffffffffffffffffffff167f7ea89a4591614515571c2b51f5ea06494056f261c10ab1ed8c03c7590d87bce08888888860405161113694939291906128ad565b60405180910390a45b5050505050505050565b333b156111d8576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603760248201527f5374616e646172644272696467653a2066756e6374696f6e2063616e206f6e6c60448201527f792062652063616c6c65642066726f6d20616e20454f410000000000000000006064820152608401610171565b610dcd868633338888888861195f565b610dcd86338787878787611728565b333014611286576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603360248201527f5374616e646172644272696467653a2066756e6374696f6e2063616e206f6e6c60448201527f792062652063616c6c65642062792073656c66000000000000000000000000006064820152608401610171565b3073ffffffffffffffffffffffffffffffffffffffff85160361132b576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602a60248201527f5374616e646172644272696467653a206c6f63616c20746f6b656e2063616e6e60448201527f6f742062652073656c66000000000000000000000000000000000000000000006064820152608401610171565b61133484611d3c565b15611482576113438484611d6e565b6113f5576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152604a60248201527f5374616e646172644272696467653a2077726f6e672072656d6f746520746f6b60448201527f656e20666f72204f7074696d69736d204d696e7461626c65204552433230206c60648201527f6f63616c20746f6b656e00000000000000000000000000000000000000000000608482015260a401610171565b6040517f40c10f1900000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff8381166004830152602482018390528516906340c10f1990604401600060405180830381600087803b15801561146557600080fd5b505af1158015611479573d6000803e3d6000fd5b50505050611504565b73ffffffffffffffffffffffffffffffffffffffff8085166000908152600260209081526040808320938716835292905220546114c09082906129ab565b73ffffffffffffffffffffffffffffffffffffffff808616600081815260026020908152604080832094891683529390529190912091909155611504908383611e15565b50505050565b6115043385348686868080601f01602080910402602001604051908101604052809392919081815260200183838082843760009201919091525061039c92505050565b7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff16633dbb202b7f0000000000000000000000000000000000000000000000000000000000000000630166a07a60e01b8a8c8b8b8b8a8a6040516024016115cf97969594939291906129c2565b604080517fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe08184030181529181526020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff167fffffffff000000000000000000000000000000000000000000000000000000009485161790525160e085901b909216825261166292918890600401612802565b600060405180830381600087803b15801561167c57600080fd5b505af1158015611690573d6000803e3d6000fd5b505050508573ffffffffffffffffffffffffffffffffffffffff168773ffffffffffffffffffffffffffffffffffffffff168973ffffffffffffffffffffffffffffffffffffffff167f7ff126db8024424bbfd9826e8ab82ff59136289ea440b04b39a0df1b03b9cabf8888878760405161113694939291906128ad565b600080600080845160208601878a8af19695505050505050565b60008773ffffffffffffffffffffffffffffffffffffffff1663c01e1bd66040518163ffffffff1660e01b8152600401602060405180830381865afa158015611775573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906117999190612847565b90507fffffffffffffffffffffffff215221522152215221522152215221522153000073ffffffffffffffffffffffffffffffffffffffff8916016118d55784341461188d576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152604360248201527f4c325374616e646172644272696467653a20455448207769746864726177616c60448201527f73206d75737420696e636c7564652073756666696369656e742045544820766160648201527f6c75650000000000000000000000000000000000000000000000000000000000608482015260a401610171565b6118d08787878787878080601f01602080910402602001604051908101604052809392919081815260200183838082843760009201919091525061039c92505050565b6118e5565b6118e5888289898989898961195f565b8673ffffffffffffffffffffffffffffffffffffffff168873ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff167f73d170910aba9e6d50b102db522b1dbcd796216f5128b445aa2135272886497e8989888860405161113694939291906128ad565b3073ffffffffffffffffffffffffffffffffffffffff891603611a04576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602a60248201527f5374616e646172644272696467653a206c6f63616c20746f6b656e2063616e6e60448201527f6f742062652073656c66000000000000000000000000000000000000000000006064820152608401610171565b611a0d88611d3c565b15611b5b57611a1c8888611d6e565b611ace576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152604a60248201527f5374616e646172644272696467653a2077726f6e672072656d6f746520746f6b60448201527f656e20666f72204f7074696d69736d204d696e7461626c65204552433230206c60648201527f6f63616c20746f6b656e00000000000000000000000000000000000000000000608482015260a401610171565b6040517f9dc29fac00000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff878116600483015260248201869052891690639dc29fac90604401600060405180830381600087803b158015611b3e57600080fd5b505af1158015611b52573d6000803e3d6000fd5b50505050611bef565b611b7d73ffffffffffffffffffffffffffffffffffffffff8916873087611ee9565b73ffffffffffffffffffffffffffffffffffffffff8089166000908152600260209081526040808320938b1683529290522054611bbb908590612a1f565b73ffffffffffffffffffffffffffffffffffffffff808a166000908152600260209081526040808320938c16835292905220555b61113f888888888888888861154d565b606081600003611c4257505060408051808201909152600181527f3000000000000000000000000000000000000000000000000000000000000000602082015290565b8160005b8115611c6c5780611c5681612a37565b9150611c659050600a83612a9e565b9150611c46565b60008167ffffffffffffffff811115611c8757611c87612ab2565b6040519080825280601f01601f191660200182016040528015611cb1576020820181803683370190505b5090505b8415611d3457611cc66001836129ab565b9150611cd3600a86612ae1565b611cde906030612a1f565b60f81b818381518110611cf357611cf3612af5565b60200101907effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916908160001a905350611d2d600a86612a9e565b9450611cb5565b949350505050565b6000611d68827f1d1d8b6300000000000000000000000000000000000000000000000000000000611f47565b92915050565b60008273ffffffffffffffffffffffffffffffffffffffff1663c01e1bd66040518163ffffffff1660e01b8152600401602060405180830381865afa158015611dbb573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190611ddf9190612847565b73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1614905092915050565b60405173ffffffffffffffffffffffffffffffffffffffff83166024820152604481018290526109af9084907fa9059cbb00000000000000000000000000000000000000000000000000000000906064015b604080517fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe08184030181529190526020810180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff167fffffffff0000000000000000000000000000000000000000000000000000000090931692909217909152611f6a565b60405173ffffffffffffffffffffffffffffffffffffffff808516602483015283166044820152606481018290526115049085907f23b872dd0000000000000000000000000000000000000000000000000000000090608401611e67565b6000611f5283612076565b8015611f635750611f6383836120da565b9392505050565b6000611fcc826040518060400160405280602081526020017f5361666545524332303a206c6f772d6c6576656c2063616c6c206661696c65648152508573ffffffffffffffffffffffffffffffffffffffff166121a59092919063ffffffff16565b8051909150156109af5780806020019051810190611fea9190612b24565b6109af576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602a60248201527f5361666545524332303a204552433230206f7065726174696f6e20646964206e60448201527f6f742073756363656564000000000000000000000000000000000000000000006064820152608401610171565b60006120a2827f01ffc9a7000000000000000000000000000000000000000000000000000000006120da565b8015611d6857506120d3827fffffffff000000000000000000000000000000000000000000000000000000006120da565b1592915050565b604080517fffffffff000000000000000000000000000000000000000000000000000000008316602480830191909152825180830390910181526044909101909152602080820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff167f01ffc9a700000000000000000000000000000000000000000000000000000000178152825160009392849283928392918391908a617530fa92503d91506000519050828015612192575060208210155b80156108d2575015159695505050505050565b6060611d3484846000858573ffffffffffffffffffffffffffffffffffffffff85163b61222e576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601d60248201527f416464726573733a2063616c6c20746f206e6f6e2d636f6e74726163740000006044820152606401610171565b6000808673ffffffffffffffffffffffffffffffffffffffff1685876040516122579190612b46565b60006040518083038185875af1925050503d8060008114612294576040519150601f19603f3d011682016040523d82523d6000602084013e612299565b606091505b50915091506108d2828286606083156122b3575081611f63565b8251156122c35782518084602001fd5b816040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610171919061261d565b73ffffffffffffffffffffffffffffffffffffffff8116811461231957600080fd5b50565b60008083601f84011261232e57600080fd5b50813567ffffffffffffffff81111561234657600080fd5b60208301915083602082850101111561235e57600080fd5b9250929050565b600080600080600080600060c0888a03121561238057600080fd5b873561238b816122f7565b9650602088013561239b816122f7565b955060408801356123ab816122f7565b945060608801356123bb816122f7565b93506080880135925060a088013567ffffffffffffffff8111156123de57600080fd5b6123ea8a828b0161231c565b989b979a50959850939692959293505050565b803563ffffffff8116811461241157600080fd5b919050565b60008060006040848603121561242b57600080fd5b612434846123fd565b9250602084013567ffffffffffffffff81111561245057600080fd5b61245c8682870161231c565b9497909650939450505050565b60008060008060006080868803121561248157600080fd5b853561248c816122f7565b9450602086013561249c816122f7565b935060408601359250606086013567ffffffffffffffff8111156124bf57600080fd5b6124cb8882890161231c565b969995985093965092949392505050565b6000806000806000608086880312156124f457600080fd5b85356124ff816122f7565b945060208601359350612514604087016123fd565b9250606086013567ffffffffffffffff8111156124bf57600080fd5b600080600080600080600060c0888a03121561254b57600080fd5b8735612556816122f7565b96506020880135612566816122f7565b95506040880135612576816122f7565b94506060880135935061258b608089016123fd565b925060a088013567ffffffffffffffff8111156123de57600080fd5b60005b838110156125c25781810151838201526020016125aa565b838111156115045750506000910152565b600081518084526125eb8160208601602086016125a7565b601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0169290920160200192915050565b602081526000611f6360208301846125d3565b60008060008060008060a0878903121561264957600080fd5b8635612654816122f7565b95506020870135612664816122f7565b945060408701359350612679606088016123fd565b9250608087013567ffffffffffffffff81111561269557600080fd5b6126a189828a0161231c565b979a9699509497509295939492505050565b600080604083850312156126c657600080fd5b82356126d1816122f7565b915060208301356126e1816122f7565b809150509250929050565b6000806000806080858703121561270257600080fd5b843561270d816122f7565b9350602085013561271d816122f7565b9250604085013561272d816122f7565b9396929550929360600135925050565b6000806000806060858703121561275357600080fd5b843561275e816122f7565b935061276c602086016123fd565b9250604085013567ffffffffffffffff81111561278857600080fd5b6127948782880161231c565b95989497509550505050565b828152604060208201526000611d3460408301846125d3565b600073ffffffffffffffffffffffffffffffffffffffff8087168352808616602084015250836040830152608060608301526127f860808301846125d3565b9695505050505050565b73ffffffffffffffffffffffffffffffffffffffff8416815260606020820152600061283160608301856125d3565b905063ffffffff83166040830152949350505050565b60006020828403121561285957600080fd5b8151611f63816122f7565b8183528181602085013750600060208284010152600060207fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0601f840116840101905092915050565b73ffffffffffffffffffffffffffffffffffffffff851681528360208201526060604082015260006127f8606083018486612864565b8381526040602082015260006128fd604083018486612864565b95945050505050565b600084516129188184602089016125a7565b80830190507f2e000000000000000000000000000000000000000000000000000000000000008082528551612954816001850160208a016125a7565b6001920191820152835161296f8160028401602088016125a7565b0160020195945050505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000828210156129bd576129bd61297c565b500390565b600073ffffffffffffffffffffffffffffffffffffffff808a1683528089166020840152808816604084015280871660608401525084608083015260c060a0830152612a1260c083018486612864565b9998505050505050505050565b60008219821115612a3257612a3261297c565b500190565b60007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8203612a6857612a6861297c565b5060010190565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b600082612aad57612aad612a6f565b500490565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b600082612af057612af0612a6f565b500690565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b600060208284031215612b3657600080fd5b81518015158114611f6357600080fd5b60008251612b588184602087016125a7565b919091019291505056fea164736f6c634300080f000a"

func init() {
	if err := json.Unmarshal([]byte(L2StandardBridgeStorageLayoutJSON), L2StandardBridgeStorageLayout); err != nil {
		panic(err)
	}

	layouts["L2StandardBridge"] = L2StandardBridgeStorageLayout
	deployedBytecodes["L2StandardBridge"] = L2StandardBridgeDeployedBin
}
