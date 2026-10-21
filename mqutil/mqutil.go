package mqutil

import (
	"stresstester/utils"
	"sync"

	// Import the v5 ibmmq package
	ibmmq "github.com/ibm-messaging/mq-golang/v5/ibmmq"
)

var queue *ibmmq.MQObject
var qMgr *ibmmq.MQQueueManager
var mqMutex sync.Mutex

func getqMgr() *ibmmq.MQQueueManager {
	return qMgr
}

func getqueue() *ibmmq.MQObject {
	return queue
}

func setqMgr(qm *ibmmq.MQQueueManager) {
	qMgr = qm
}

func setqueue(q *ibmmq.MQObject) {
	queue = q
}

func Init(channelname, connectionname, qmgrName, qName, authenticationType, userid, password string) error {

	//os.Setenv("MQSERVER", Connectionstring)

	//log.Printf("Attempting connection to Queue Manager: %s\n", qMgrName)

	cno := ibmmq.NewMQCNO()

	// Direct client routing flag forces MQ to fetch parameters out of the OS environment.
	cno.Options = ibmmq.MQCNO_CLIENT_BINDING

	cd := ibmmq.NewMQCD()
	cd.ChannelName = channelname       //"DEV.APP.SVRCONN"
	cd.ConnectionName = connectionname //"localhost(1414)" // Adjust to your container IP/Port if different
	cno.ClientConn = cd

	var csp *ibmmq.MQCSP

	switch authenticationType {
	case "MQCSP_AUTH_USER_ID_AND_PWD":
		csp = ibmmq.NewMQCSP()
		csp.AuthenticationType = ibmmq.MQCSP_AUTH_USER_ID_AND_PWD
		csp.AuthenticationType = ibmmq.MQCSP_AUTH_USER_ID_AND_PWD
		csp.UserId = userid //"app"
		decpasssword, err := utils.Decodebase64(password)
		if err != nil {
			csp.Password = password //"passw0rd"
		} else {
			csp.Password = decpasssword //"passw0rd"
		}
		cno.SecurityParms = csp
	default:

	}

	// 3. CONNECT TO THE QUEUE MANAGER
	qMgr, err := ibmmq.Connx(qmgrName, cno)
	if err != nil {
		return err
	}
	setqMgr(&qMgr)

	// 4. SET OPEN OPTIONS FOR OUTPUT (SENDING)
	od := ibmmq.NewMQOD()
	od.ObjectName = qName
	openOptions := ibmmq.MQOO_OUTPUT

	// 5. OPEN THE QUEUE
	queue, err := qMgr.Open(od, openOptions)
	if err != nil {
		return err
	}
	setqueue(&queue)

	return nil
}

func CloseConn() {
	defer qMgr.Disc()
	defer queue.Close(0) // Ensure queue closes on exit
}

func Putmessage(message string) error {
	// 6. INITIALIZE MESSAGE DESCRIPTOR AND PUT OPTIONS
	mqMutex.Lock()
	defer mqMutex.Unlock()

	md := ibmmq.NewMQMD()
	pmo := ibmmq.NewMQPMO()
	pmo.Options = ibmmq.MQPMO_NO_SYNCPOINT // Send outside of a transaction unit

	// Set message format as string/text data
	md.Format = ibmmq.MQFMT_STRING

	// 7. DEFINE PAYLOAD AND TRANSMIT
	messagePayload := message
	buffer := []byte(messagePayload)

	err := queue.Put(md, pmo, buffer)
	if err != nil {
		return err
	}
	return nil
}
