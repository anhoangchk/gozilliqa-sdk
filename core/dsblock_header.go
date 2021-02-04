package core

import (
	"github.com/Zilliqa/gozilliqa-sdk/protobuf"
	"github.com/Zilliqa/gozilliqa-sdk/util"
	"github.com/golang/protobuf/proto"
	"math/big"
	"strconv"
)

// https://github.com/Zilliqa/Zilliqa/blob/04162ef0c3c1787ebbd940b7abd6b7ff1d4714ed/src/libData/BlockData/BlockHeader/DSBlockHeader.h
type DsBlockHeader struct {
	blockHeaderBase BlockHeaderBase
	DsDifficulty    uint32
	Difficulty      uint32
	// The one who proposed this DS block
	// base16 string
	LeaderPubKey string
	// Block index, starting from 0 in the genesis block
	BlockNum uint64
	// Tx Epoch Num then the DS block was generated
	EpochNum uint64
	GasPrice string
	SwInfo       SWInfo
	// key is (base16) public key
	PoWDSWinners map[string]Peer
	// (base16) public key
	RemoveDSNodePubKeys []string
	// todo concrete data type
	dSBlockHashSet     DSBlockHashSet
	GovDSShardVotesMap map[uint32]Pair
}

func NewDsBlockHeaderFromDsBlockT(dst *DsBlockT) *DsBlockHeader {
	dsBlockHeader := &DsBlockHeader{}
	dsBlockHeader.DsDifficulty = dst.Header.DifficultyDS
	dsBlockHeader.Difficulty = dst.Header.Difficulty
	dsBlockHeader.LeaderPubKey = dst.Header.LeaderPubKey

	blockNum, _ := strconv.ParseUint(dst.Header.BlockNum, 10, 64)
	dsBlockHeader.BlockNum = blockNum

	epochNum, _ := strconv.ParseUint(dst.Header.EpochNum, 10, 64)
	dsBlockHeader.EpochNum = epochNum

	dsBlockHeader.GasPrice = dst.Header.GasPrice

	zilliqaUpgradeDS, _ := strconv.ParseUint(dst.Header.SWInfo.Zilliqa[3].(string), 10, 64)
	scillaUpgradeDS, _ := strconv.ParseUint(dst.Header.SWInfo.Scilla[3].(string), 10, 64)

	dsBlockHeader.SwInfo = SWInfo{
		ZilliqaMajorVersion: uint32(dst.Header.SWInfo.Zilliqa[0].(float64)),
		ZilliqaMinorVersion: uint32(dst.Header.SWInfo.Zilliqa[1].(float64)),
		ZilliqaFixVersion:   uint32(dst.Header.SWInfo.Zilliqa[2].(float64)),
		ZilliqaUpgradeDS:    zilliqaUpgradeDS,
		ZilliqaCommit:       uint32(dst.Header.SWInfo.Zilliqa[4].(float64)),
		ScillaMajorVersion:  uint32(dst.Header.SWInfo.Scilla[0].(float64)),
		ScillaMinorVersion:  uint32(dst.Header.SWInfo.Scilla[1].(float64)),
		ScillaFixVersion:    uint32(dst.Header.SWInfo.Scilla[2].(float64)),
		ScillaUpgradeDS:     scillaUpgradeDS,
		ScillaCommit:        uint32(dst.Header.SWInfo.Scilla[4].(float64)),
	}

	winnermap := make(map[string]Peer, len(dst.Header.PoWWinners))
	for i := 0; i < len(dst.Header.PoWWinners); i++ {
		ip := dst.Header.PoWWinnersIP[i].IP
		port := dst.Header.PoWWinnersIP[i].Port

		IPAddress := IP2Long(ip)

		peer := Peer{
			IpAddress:      new(big.Int).SetUint64(uint64(IPAddress)),
			ListenPortHost: port,
		}
		winnermap[dst.Header.PoWWinners[i]] = peer
	}

	dsBlockHeader.PoWDSWinners = winnermap

	var removeDSNodePubKeys []string
	for _, key := range dst.Header.MembersEjected {
		removeDSNodePubKeys = append(removeDSNodePubKeys, key)
	}
	dsBlockHeader.RemoveDSNodePubKeys = removeDSNodePubKeys

	var dsHashSet DSBlockHashSet
	dsHashSet.shadingHash = util.DecodeHex(dst.Header.ShardingHash)
	dsBlockHeader.dSBlockHashSet = dsHashSet

	governance := make(map[uint32]Pair,0)
	govs := dst.Header.Governance
	for _, gov := range govs {
		proposalId := gov.ProposalId
		dsmap := make(map[uint32]uint32,0)
		dsvotes := gov.DSVotes
		for _,dsvote := range dsvotes {
			dsmap[dsvote.VoteValue] = dsvote.VoteCount
		}

		shardmap := make(map[uint32]uint32,0)
		shardvotes := gov.ShardVotes
		for _,shardvote := range shardvotes {
			shardmap[shardvote.VoteValue] = shardvote.VoteCount
		}

		pair := Pair{
			First:  dsmap,
			Second: shardmap,
		}
		governance[proposalId] = pair
	}

	dsBlockHeader.GovDSShardVotesMap = governance


	dsBlockHeader.blockHeaderBase.Version = dst.Header.Version
	ch := util.DecodeHex(dst.Header.CommitteeHash)
	var commitHash [32]byte
	copy(commitHash[:], ch)
	dsBlockHeader.blockHeaderBase.CommitteeHash = commitHash

	ph := util.DecodeHex(dst.Header.PrevHash)
	var prevHash [32]byte
	copy(prevHash[:], ph)
	dsBlockHeader.blockHeaderBase.PrevHash = prevHash

	return dsBlockHeader
}

func (d *DsBlockHeader) Serialize() []byte {
	h := d.ToProtobuf(false)
	bytes, _ := proto.Marshal(h)
	return bytes
}

// the default value of concreteVarsOnly should be false
func (d *DsBlockHeader) ToProtobuf(concreteVarsOnly bool) *protobuf.ProtoDSBlock_DSBlockHeader {
	protoDSBlockHeader := &protobuf.ProtoDSBlock_DSBlockHeader{}
	protoBlockHeaderBase := d.blockHeaderBase.ToProtobuf()
	protoDSBlockHeader.Blockheaderbase = protoBlockHeaderBase

	if !concreteVarsOnly {
		protoDSBlockHeader.Dsdifficulty = d.DsDifficulty
		protoDSBlockHeader.Difficulty = d.Difficulty
		data := make([]byte, 0)
		gasPriceInt, _ := new(big.Int).SetString(d.GasPrice, 10)
		data = UintToByteArray(data, 0, gasPriceInt, 16)
		protoDSBlockHeader.Gasprice = &protobuf.ByteArray{
			Data: data,
		}

		var protobufWinners []*protobuf.ProtoDSBlock_DSBlockHeader_PowDSWinners
		for key, winner := range d.PoWDSWinners {
			protobufWinner := &protobuf.ProtoDSBlock_DSBlockHeader_PowDSWinners{
				Key: &protobuf.ByteArray{Data: util.DecodeHex(key)},
				Val: &protobuf.ByteArray{Data: winner.Serialize()},
			}
			protobufWinners = append(protobufWinners, protobufWinner)
		}
		protoDSBlockHeader.Dswinners = protobufWinners

		var proposals []*protobuf.ProtoDSBlock_DSBlockHeader_Proposal
		for proposal, pair := range d.GovDSShardVotesMap {
			protoproposal := &protobuf.ProtoDSBlock_DSBlockHeader_Proposal{}
			protoproposal.Proposalid = proposal

			var dsvotes []*protobuf.ProtoDSBlock_DSBlockHeader_Vote
			for value, count := range pair.First {
				dsvote := &protobuf.ProtoDSBlock_DSBlockHeader_Vote{
					Value: value,
					Count: count,
				}
				dsvotes = append(dsvotes, dsvote)
			}

			var minerVotes []*protobuf.ProtoDSBlock_DSBlockHeader_Vote
			for value, count := range pair.Second {
				minerVote := &protobuf.ProtoDSBlock_DSBlockHeader_Vote{
					Value: value,
					Count: count,
				}
				minerVotes = append(minerVotes, minerVote)
			}

			protoproposal.Dsvotes = dsvotes
			protoproposal.Minervotes = minerVotes
			proposals = append(proposals, protoproposal)
		}


		protoDSBlockHeader.Proposals = proposals

		var dsremoved []*protobuf.ByteArray
		for _, key := range d.RemoveDSNodePubKeys {
			dr := &protobuf.ByteArray{
				Data: util.DecodeHex(key),
			}
			dsremoved = append(dsremoved, dr)
		}
		protoDSBlockHeader.Dsremoved = dsremoved
	}

	protoDSBlockHeader.Leaderpubkey = &protobuf.ByteArray{Data: util.DecodeHex(d.LeaderPubKey)}
	protoDSBlockHeader.Blocknum = d.BlockNum
	protoDSBlockHeader.Epochnum = d.EpochNum

	protoDSBlockHeader.Swinfo = &protobuf.ByteArray{Data: d.SwInfo.Serialize()}

	hashset := &protobuf.ProtoDSBlock_DSBlockHashSet{
		Shardinghash:  d.dSBlockHashSet.shadingHash,
		Reservedfield: d.dSBlockHashSet.reservedField[:],
	}
	protoDSBlockHeader.Hash = hashset

	return protoDSBlockHeader
}
