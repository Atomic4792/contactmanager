package RedisConnector

import (
	config "../Config"
	Log "../Logger"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-redis/redis"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var ConfigData = config.Params{}

const hashSlots = 16384

/**
 * Pipelinedata aggregates data collection for
 * usage with pipeline bursting
 */
type PipelineData struct {
	nodeData          map[string]map[string]int
	lastBurstResults  map[string]int64
	burstReady        bool
	lastExecutionTime time.Time
	currentNodeSize   uint64
}

/**
 * Cluster Environment. Contains pointers to master
 * redis clients and slave redis clients by the
 * node id. Also holds a mutex for locking and unlocking of
 * retrieval of redis clients in usage of possible
 * concurrency (not implemented). This would be to
 * ensure the correct client is received.
 */
type ClusterScenario struct {
	masterNodes map[string]*RedisMasterNode
	slaveNodes  map[string]*RedisSlaveNode
	mu          sync.Mutex
}

/**
 * Master node structure
 */
type RedisMasterNode struct {
	id            string
	startHashSlot int
	endHashSlot   int
	address       string
	host          string
	port          string
	client        *redis.Client
}

/**
 * Slave node structure
 */
type RedisSlaveNode struct {
	id      string
	master  string
	address string
	host    string
	port    string
	client  *redis.Client
}

/**
 * JSON Configuration Data structure
 */
type ConfigurationData struct {
	GossipClientPort string `json:"gossipClientPort"`
	GossipClientHost string `json:"gossipClientHost"`
}

/**
 * Gossip Client. Initiates the node mappings and initial
 * cluster operations (CLUSTER MEET and CLUSTER REPLICATE)
 */
var gossipClient *redis.Client

/**
 * Read write lock for configuration handling
 */
var configLock = new(sync.RWMutex)

/**
 * Configuration Data structure
 */
// var config *ConfigurationData

/**
 * Cluster Scenario Structure
 */
var clusterScenario *ClusterScenario

/**
 * Pipeline Data
 */
var p *PipelineData

/**
 * CRC16 Mappings
 */
var crc16tab = [256]uint16{
	0x0000, 0x1021, 0x2042, 0x3063, 0x4084, 0x50a5, 0x60c6, 0x70e7,
	0x8108, 0x9129, 0xa14a, 0xb16b, 0xc18c, 0xd1ad, 0xe1ce, 0xf1ef,
	0x1231, 0x0210, 0x3273, 0x2252, 0x52b5, 0x4294, 0x72f7, 0x62d6,
	0x9339, 0x8318, 0xb37b, 0xa35a, 0xd3bd, 0xc39c, 0xf3ff, 0xe3de,
	0x2462, 0x3443, 0x0420, 0x1401, 0x64e6, 0x74c7, 0x44a4, 0x5485,
	0xa56a, 0xb54b, 0x8528, 0x9509, 0xe5ee, 0xf5cf, 0xc5ac, 0xd58d,
	0x3653, 0x2672, 0x1611, 0x0630, 0x76d7, 0x66f6, 0x5695, 0x46b4,
	0xb75b, 0xa77a, 0x9719, 0x8738, 0xf7df, 0xe7fe, 0xd79d, 0xc7bc,
	0x48c4, 0x58e5, 0x6886, 0x78a7, 0x0840, 0x1861, 0x2802, 0x3823,
	0xc9cc, 0xd9ed, 0xe98e, 0xf9af, 0x8948, 0x9969, 0xa90a, 0xb92b,
	0x5af5, 0x4ad4, 0x7ab7, 0x6a96, 0x1a71, 0x0a50, 0x3a33, 0x2a12,
	0xdbfd, 0xcbdc, 0xfbbf, 0xeb9e, 0x9b79, 0x8b58, 0xbb3b, 0xab1a,
	0x6ca6, 0x7c87, 0x4ce4, 0x5cc5, 0x2c22, 0x3c03, 0x0c60, 0x1c41,
	0xedae, 0xfd8f, 0xcdec, 0xddcd, 0xad2a, 0xbd0b, 0x8d68, 0x9d49,
	0x7e97, 0x6eb6, 0x5ed5, 0x4ef4, 0x3e13, 0x2e32, 0x1e51, 0x0e70,
	0xff9f, 0xefbe, 0xdfdd, 0xcffc, 0xbf1b, 0xaf3a, 0x9f59, 0x8f78,
	0x9188, 0x81a9, 0xb1ca, 0xa1eb, 0xd10c, 0xc12d, 0xf14e, 0xe16f,
	0x1080, 0x00a1, 0x30c2, 0x20e3, 0x5004, 0x4025, 0x7046, 0x6067,
	0x83b9, 0x9398, 0xa3fb, 0xb3da, 0xc33d, 0xd31c, 0xe37f, 0xf35e,
	0x02b1, 0x1290, 0x22f3, 0x32d2, 0x4235, 0x5214, 0x6277, 0x7256,
	0xb5ea, 0xa5cb, 0x95a8, 0x8589, 0xf56e, 0xe54f, 0xd52c, 0xc50d,
	0x34e2, 0x24c3, 0x14a0, 0x0481, 0x7466, 0x6447, 0x5424, 0x4405,
	0xa7db, 0xb7fa, 0x8799, 0x97b8, 0xe75f, 0xf77e, 0xc71d, 0xd73c,
	0x26d3, 0x36f2, 0x0691, 0x16b0, 0x6657, 0x7676, 0x4615, 0x5634,
	0xd94c, 0xc96d, 0xf90e, 0xe92f, 0x99c8, 0x89e9, 0xb98a, 0xa9ab,
	0x5844, 0x4865, 0x7806, 0x6827, 0x18c0, 0x08e1, 0x3882, 0x28a3,
	0xcb7d, 0xdb5c, 0xeb3f, 0xfb1e, 0x8bf9, 0x9bd8, 0xabbb, 0xbb9a,
	0x4a75, 0x5a54, 0x6a37, 0x7a16, 0x0af1, 0x1ad0, 0x2ab3, 0x3a92,
	0xfd2e, 0xed0f, 0xdd6c, 0xcd4d, 0xbdaa, 0xad8b, 0x9de8, 0x8dc9,
	0x7c26, 0x6c07, 0x5c64, 0x4c45, 0x3ca2, 0x2c83, 0x1ce0, 0x0cc1,
	0xef1f, 0xff3e, 0xcf5d, 0xdf7c, 0xaf9b, 0xbfba, 0x8fd9, 0x9ff8,
	0x6e17, 0x7e36, 0x4e55, 0x5e74, 0x2e93, 0x3eb2, 0x0ed1, 0x1ef0,
}

/**
 * Returns the node slot assigned to a hash slot.
 * Hash slots are determined by crc16 calculation
 * of the query key
 */
func (c *ClusterScenario) GetNodeSlotByHashSlot(key string) string {
	hashSlot := HashSlot(key)
	c.mu.Lock()
	var redisNodeId string
	defer c.mu.Unlock()
	for _, redisNode := range c.masterNodes {
		if redisNode.startHashSlot <= hashSlot && redisNode.endHashSlot >= hashSlot {
			redisNodeId = redisNode.id
		}
	}
	return redisNodeId
}

/**
 * Adds data to a node slot for future node pipeline bursting.
 * Updates the node size integer for fast access querying
 * in determination of ready bursting
 */
func (p *PipelineData) addNodeData(key string) {
	nodeSlot := clusterScenario.GetNodeSlotByHashSlot(key)
	i := p.nodeData[nodeSlot][key]
	i++
	p.nodeData[nodeSlot][key] = i
	j := p.currentNodeSize
	j++
	p.currentNodeSize = j
	if p.currentNodeSize == 10 {
		p.Burst()
	}
}

/**
 * Clears node data from the pipeline structure
 */
func (p *PipelineData) clearNodeData() {
	p.nodeData = make(map[string]map[string]int, 0)
	for i, _ := range clusterScenario.masterNodes {
		p.nodeData[i] = make(map[string]int, 0)
	}
}

func (p *PipelineData) Burst() {
	p.burstReady = false
	data := p.nodeData
	result := map[string]*redis.IntCmd{}
	for v := range data {
		client := clusterScenario.GetConn(v)
		pipe := client.Pipeline()
		for i, x := range p.nodeData[v] {
			j := int64(x)
			result[v+"."+i] = pipe.IncrBy(i, j)
		}
		_, err := pipe.Exec()
		CheckErrorf(err, "Pipeline Issue")
	}
	res2 := map[string]int64{}
	for k, v := range result {
		res2[k] = v.Val()
	}
	p.lastExecutionTime = time.Now()
	p.lastBurstResults = res2
	p.clearNodeData()
	p.currentNodeSize = 0
	p.burstReady = true
}

/**
 * Counts the number of failure reports for a specific node. Useful
 * in determining an agreed FAIL STATE by the cluster masters
 */
func (c *ClusterScenario) CountNodeFailureReports(nodeSlot string) {

}

/**
 * Counts the number of keys in a hash slot
 */
func (c *ClusterScenario) CountKeysInSlot(hashSlot int) {

}

/**
 * Returns cluster information
 */
func (c *ClusterScenario) GetClusterInfo() {

}

/**
 * Add a master node to the cluster scenario master nodes map
 */
func (rn *ClusterScenario) AddMasterNode(id string, node *RedisMasterNode) {
	rn.masterNodes[id] = node
}

/**
 * Add a slave node to the cluster scenario slave nodes map
 */
func (rn *ClusterScenario) AddSlaveNode(id string, node *RedisSlaveNode) {
	rn.slaveNodes[id] = node
}

/**
 * Master node to string
 */
func (rn *RedisMasterNode) ToString(prefix string, full bool) string {
	if full {
		return fmt.Sprintf(prefix+" %s %s %d-%d",
			rn.address, rn.id, rn.startHashSlot, rn.endHashSlot)
	} else {
		return fmt.Sprintf(prefix+" %s %s",
			rn.address, rn.id)
	}
}

/**
 * Slave node to string
 */
func (rn *RedisSlaveNode) ToString(prefix string) string {
	return fmt.Sprintf(prefix+" %s %s %s",
		rn.address, rn.id, rn.master)
}

/**
 * Greet a master node. Performs a CLUSTER MEET command
 * against the node which was aggregated from the initiating
 * CLUSTER NODES parsing. This command gets called on initiation
 * of the script as well as when a node is MOVED as the build mappings
 * command gets recalled during that response. We only have to
 * execute this command once with the originating greeting client
 */
func (rn *RedisMasterNode) Greet(gossip *redis.Client) {
	err := gossip.ClusterMeet(rn.host, rn.port).Err()
	CheckErrorf(err, "Gossip Greet Error "+rn.address)
	//fmt.Println(fmt.Sprintf("Greetings Master %s %s (HashSlots: %d-%d)",
	//	rn.address, rn.id, rn.startHashSlot, rn.endHashSlot))
}

/**
 * Greet a slave node. Performs a CLUSTER MEET command
 * against the node which was aggregated from the initiating
 * CLUSTER NODES parsing. This command gets called on initiation
 * of the script as well as when a node is MOVED as the build mappings
 * command gets recalled during that response. We only have to
 * execute this command once with the originating greeting client
 */
func (rn *RedisSlaveNode) Greet(gossip *redis.Client) {
	err := gossip.ClusterMeet(rn.host, rn.port).Err()
	CheckErrorf(err, "Gossip Greet Error "+rn.address)
	//fmt.Println(fmt.Sprintf("Greetings Slave %s %s (Master: %s)",
	//	rn.address, rn.id, rn.master))
}

/**
 * Returns a master node redis client
 */
func (rn *ClusterScenario) GetConn(id string) *redis.Client {
	return rn.masterNodes[id].client
}

/**
 * Sends a CLUSTER REPLICATE command from a slave node.
 */
func (rn *RedisSlaveNode) Replicate() {
	rn.client.ClusterReplicate(rn.master)
	//fmt.Println(fmt.Sprintf("Replicated Slave %s %s (Master: %s)",
	//	rn.address, rn.id, rn.master))
}

/**
 * Error handling function
 */
func CheckErrorf(err error, message string) {
	if err != nil {
		Log.LogMsg(5, message+" "+err.Error())
	}
}

/**
 * Computes a CRC16 integer based off string
 */
func crc16(buf string) uint16 {
	var crc uint16
	for _, n := range buf {
		crc = (crc << uint16(8)) ^ crc16tab[((crc>>uint16(8))^uint16(n))&0x00FF]
	}
	return crc
}

/**
 * Determines hash slot storage integer based off query string.
 */
func HashSlot(key string) int {
	if start := strings.Index(key, "{"); start >= 0 {
		// if end == 0, then it's {}, so we ignore it
		if end := strings.Index(key[start+1:], "}"); end > 0 {
			end += start + 1
			key = key[start+1 : end]
		}
	}
	return int(crc16(key) % hashSlots)
}

/**
 * Dump utility
 */
func Dump(a ...interface{}) {
	spew.Dump(a...)
}

/**
 * Handles query response. For Moved, we initiate a rebuild
 * of node mappings.
 */
func HandleResponse(resp string) {
	if strings.Contains(resp, "MOVED") {
		//fmt.Println("MOVED RESPONSE")
		//BuildNodeMappings()
	} else if strings.Contains(resp, "OK") {
		//fmt.Println("SUCCESS RESPONSE")
	}
}

/**
 * Cluster Scenario perform redis get command wrapper
 */
func (cs *ClusterScenario) DoGet(query string) interface{} {
	return GetCmd(query)
}

/**
 * Cluster Scenario perform redis set command wrapper
 */
func (cs *ClusterScenario) DoSet(query string, val interface{}) {
	SetCmd(query, val)
}

func TestSetRedirect() {
	client := GetRedisClientAdapter(5000)
	val := client.Set("{customer}.name", "bar1", 0).String()
	HandleResponse(val)
}

/**
 * Internal redis get cmd
 */
func GetCmd(queryString string) interface{} {
	hashSlot := HashSlot(queryString)
	redisClient := GetRedisClientAdapter(hashSlot)
	val, err := redisClient.Get(queryString).Result()
	if err == redis.Nil {
		CheckErrorf(err, fmt.Sprintf("Key %s Does Not Exist", queryString))
	}
	CheckErrorf(err, "Redis Get Error")
	return val
}

///****
//* Internal Redis command for zadd
//*/
//
//func Zadd(key string, score string, member string) interface{}{
//	hashSlot := HashSlot(key)
//	redisClient := GetRedisClientAdapter(hashSlot)
//	val , err := redisClient.ZAdd(key, score + " " + member)
//	if err != nil {
//		return err
//	}
//
//	return val
//
//}

/**
 * Internal redis set cmd
 */
func SetCmd(queryString string, val interface{}) {
	hashSlot := HashSlot(queryString)
	redisClient := GetRedisClientAdapter(hashSlot)
	resp := redisClient.Set(queryString, val, 0).String()
	HandleResponse(resp)
}

/**
 * Get a master redis client based on hash slot which fetches
 * from node structure stored in cluster scenario
 */
func GetRedisClientAdapter(hashSlot int) *redis.Client {
	clusterScenario.mu.Lock()
	defer clusterScenario.mu.Unlock()
	for _, redisNode := range clusterScenario.masterNodes {
		if redisNode.startHashSlot <= hashSlot && redisNode.endHashSlot >= hashSlot {
			//fmt.Println(redisNode.ToString("Client Requested", false))
			return redisNode.client
		}
	}
	return nil
}

/**
 * Start a redis client
 */
func startRedisClient(address string, password string, db int) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       db,
	})
	_, err := client.Ping().Result()
	if err != nil {
		// LogRedisConnError("Could not ping " + address, err)
		Log.LogMsg(5, "Could not ping "+address+" "+err.Error())
	}
	return client
}

/**
 * initiates the cluster scenario, zero values the master/slave node maps.
 * Starts a gossip client based on JSON config. we use this client to call
 * CLUSTER NODES and create our node mappings, including master and slave redis
 * clients. Then, we query CLUSTER MEET commands with the gossip client against
 * the populated mappings. This might be unuseful in initial mappings but for remappings
 * during a move this might be useful for fast initiation. We only have to query
 * each node once with the original gossip client. In the case of a remapping,
 * such as with a MOVE detect, we call this function which should then immediately
 * initiate an internal gossip. This code is functional, but the CLUSTER MEET has not been
 * thoroughly tested. The end goal would be to have a fast rebuild of the gossip
 * channel when a MOVE is detected. Also, i call CLUSTER REPLICATE on the
 * slave nodes, which might not be neccessaray however also for fast rebuild
 * would be needed.
 */
func BuildNodeMappings() {
	// start cluster scenario
	clusterScenario = &ClusterScenario{
		masterNodes: make(map[string]*RedisMasterNode),
		slaveNodes:  make(map[string]*RedisSlaveNode),
	}

	// load gossip master
	gossipAddr := net.JoinHostPort(ConfigData.Redis.Host, ConfigData.Redis.Port)
	gossipClient := startRedisClient(gossipAddr, "", 0)

	// load node mapping hashes
	masterNodes := make(map[int][]string, 0)
	slaveNodes := make(map[int][]string, 0)

	// translate nodes
	nodes, err := gossipClient.ClusterNodes().Result()
	CheckErrorf(err, "Cluster Nodes Error")
	s := strings.Split(nodes, "\n")
	var j, x int
	for _, node := range s {
		vs := strings.Split(node, " ")
		for _, n := range vs {
			if strings.Contains(n, "master") {
				masterNodes[j] = vs
				j++
			} else if strings.Contains(n, "slave") {
				slaveNodes[x] = vs
				x++
			}
		}
	}

	// build redis master nodes
	for _, n := range masterNodes {
		nodeAddr := strings.Split(n[1], "@")[0]
		slots := strings.Split(n[8], "-")
		var nodeClient *redis.Client
		if gossipAddr == nodeAddr {
			nodeClient = gossipClient
		} else {
			nodeClient = startRedisClient(nodeAddr, "", 0)
		}
		startHashSlot, err := strconv.Atoi(slots[0])
		CheckErrorf(err, "Convert Start Hash Slot Error")
		endHashSlot, err := strconv.Atoi(slots[1])
		CheckErrorf(err, "Convert End Hash Slot Error")
		addrParts := strings.Split(nodeAddr, ":")
		CheckErrorf(err, "Convert Port Error")
		redisNode := RedisMasterNode{
			id:            n[0],
			client:        nodeClient,
			startHashSlot: startHashSlot,
			endHashSlot:   endHashSlot,
			address:       nodeAddr,
			host:          addrParts[0],
			port:          addrParts[1],
		}
		redisNode.Greet(gossipClient)
		clusterScenario.AddMasterNode(n[0], &redisNode)
	}

	// build redis slave nodes
	for _, n := range slaveNodes {
		nodeAddr := strings.Split(n[1], "@")[0]
		nodeClient := startRedisClient(nodeAddr, "", 0)
		addrParts := strings.Split(nodeAddr, ":")
		CheckErrorf(err, "Convert Port Error")
		redisNode := RedisSlaveNode{
			id:      n[0],
			client:  nodeClient,
			address: nodeAddr,
			host:    addrParts[0],
			port:    addrParts[1],
			master:  n[3],
		}
		redisNode.Greet(gossipClient)
		redisNode.Replicate()
		clusterScenario.AddSlaveNode(n[0], &redisNode)
	}

	// start pipeline data
	p = &PipelineData{
		nodeData:          make(map[string]map[string]int, 0),
		lastBurstResults:  make(map[string]int64, 0),
		burstReady:        true,
		lastExecutionTime: time.Now(),
		currentNodeSize:   0,
	}
	for i, _ := range clusterScenario.masterNodes {
		p.nodeData[i] = make(map[string]int, 0)
	}
}

/**
 * Initializes configuration data. Builds node mappings.
 * opens a USR2 signal which makes possible for
 * realtime reloading of configuration data and mappings
 * without restarting the program IF NEEDED.
 * You would call as:
 * killall -s USR2 {programname} (such as killall -s USR2 thisProgramName)
 */
func init() {
	// LoadConfig()
	BuildNodeMappings()
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGUSR2)
	go func() {
		for {
			<-s
			// LoadConfig()
			clusterScenario = nil
			BuildNodeMappings()
		}
	}()
}

func AddRedisData(key string) {
	p.addNodeData(key)
}

func FlushRedisPipelineData() {
	p.Burst()
}

func FetchLastPipelineBurstResults() map[string]int64 {
	return p.lastBurstResults
}

func GetRedisClientByKey(key string) *redis.Client {

	return GetRedisClientAdapter(HashSlot(key))
}

/**
 * Pipeline test bank
 */

/**
 * Standard test bank
 */
//func StandardTest() {
//	fmt.Println()
//	hashSlotA := HashSlot("{customer}.name")
//	fmt.Println("TEST MOVED DETECTED {customer}.name ===> HASH SLOT ", hashSlotA)
//	TestSetRedirect()
//	fmt.Println("*******************")
//	fmt.Println()
//	hashSlotB := HashSlot("{cust}.name")
//	fmt.Println("TEST GET CLUSTER QUERY {cust}.name ===> HASH SLOT ", hashSlotB)
//	val := clusterScenario.DoGet("{cust}.name")
//	fmt.Println("VALUE => ", val)
//	fmt.Println()
//	hashSlotC := HashSlot("{custTEST}.name")
//	fmt.Println("TEST SET CLUSTER QUERY {custTEST}.name WITH testdata ===> HASH SLOT ", hashSlotC)
//	clusterScenario.DoSet("{custTEST}.name", "testdata")
//	val = clusterScenario.DoGet("{custTEST}.name")
//	fmt.Println("VALUE RETRIEVED ==> ", val)
//	fmt.Println()
//}
