/*******************************************************************************
 * Copyright 2020 Samsung Electronics All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 *******************************************************************************/

package server

import (
	"net"
)

//clientConnection structure for client
type clientConnection struct {
	conn            net.Conn
	server          *Server
	outgoingChannel chan *NetPacket
	localAddr       string
	isConnected     bool
	deviceID        string
}

func (c *clientConnection) initClient(s *Server) {

	logPrefix := logTag + "[initClient]"
	c.outgoingChannel = make(chan *NetPacket, channelSize)
	c.server = s

	log.Printf("%s New connection from %s (%s) initialised\n", logPrefix, c.conn.RemoteAddr().String(), c.deviceID)
	c.localAddr = c.conn.RemoteAddr().String()
	c.isConnected = true
	go c.recvRoutine(s.incomingIPPacketChan)
	go c.sendRoutine()
}

//reads from the cleint's connection
func (c *clientConnection) recvRoutine(sink chan *NetPacketIP) {
	logPrefix := logTag + "[ClientRecvRoutine]"

	vpnbuf := make([]byte, packetSize)
	for c.isConnected && c.server.isAlive {
		n, err := c.conn.Read(vpnbuf)
		if err != nil {
			log.Printf("%s Could not Read from c.conn: %s", logPrefix, err.Error())
			c.hadError(false)
			return
		}
		if n > 0 && err == nil {
			localAddr := c.conn.RemoteAddr()
			c.server.SetClientAddress(c.deviceID, localAddr.String())
			sink <- &NetPacketIP{Packet: &NetPacket{Packet: vpnbuf}, ClientID: c.deviceID}
		}
	}
}

//writes to the client connection
func (c *clientConnection) sendRoutine() {
	logPrefix := logTag + "[ClientSendRoutine]"
	for c.isConnected && c.server.isAlive {
		select {
		case pkt := <-c.outgoingChannel:

			_, err := c.conn.Write(pkt.Packet)
			if err != nil {
				log.Println(logPrefix, "Write error for", c.conn.RemoteAddr().String(), err.Error())
				c.hadError(false)
				return
			}
		}
	}
}

func (c *clientConnection) hadError(errInRead bool) {
	logPrefix := logTag + "[hadError]"
	if !errInRead {
		c.conn.Close()
	}
	c.isConnected = false
	c.server.RemoveClient(c.deviceID)
	log.Println(logPrefix, "client id:", c.deviceID, "closed")
}

func (c *clientConnection) queueIP(pkt *NetPacket) {
	logPrefix := logTag + "[queueIP]"
	select {
	case c.outgoingChannel <- pkt:
		//log.Println(logPrefix, "Written on client's outgoingChannel")
	default:
		log.Println(logPrefix, "Warning: Dropping packets for", c.conn.RemoteAddr().String(), "as outbound msg queue is full")
	}
}
