package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

/*** UTIL ***/

func debug(values ...any) {
	fmt.Fprintln(os.Stderr, values...)
}

/*** MAP ***/

type Node struct {
	index                int
	isBobnetAgentPresent bool
	isExit               bool
	links                []*Node
}

func (node *Node) String() string {
	if node == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(strconv.FormatInt(int64(node.index), 10))
	if node.isExit {
		sb.WriteString(">")
	}
	if node.isBobnetAgentPresent {
		sb.WriteString("*")
	}
	sb.WriteString("(")
	for i := 0; i < len(node.links); i++ {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(strconv.FormatInt(int64(node.links[i].index), 10))
	}
	sb.WriteString(")")
	return sb.String()
}

func (node *Node) removeLink(linkedNode *Node) {
	if node == nil || linkedNode == nil {
		return
	}
	linkedNodeIndex := -1
	for i, n := range node.links {
		if n.index == linkedNode.index {
			linkedNodeIndex = i
			break
		}
	}
	if linkedNodeIndex == -1 {
		return
	}
	if len(node.links) == 1 {
		node.links = []*Node{}
		return
	}
	tmp := node.links[0]
	node.links[0] = node.links[linkedNodeIndex]
	node.links[linkedNodeIndex] = tmp
	node.links = node.links[1:]
}

type Link struct {
	node1 *Node
	node2 *Node
}

func (link *Link) String() string {
	if link == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("(")
	if link.node1 != nil {
		sb.WriteString(strconv.FormatInt(int64(link.node1.index), 10))
	}
	if link.node2 != nil {
		if link.node1 != nil {
			sb.WriteString(",")
		}
		sb.WriteString(strconv.FormatInt(int64(link.node2.index), 10))
	}
	sb.WriteString(")")
	return sb.String()
}

type GameMap struct {
	bobnetAgentIndex int
	exits            []*Node
	links            []*Link
	nodes            []*Node
}

func (gameMap *GameMap) String() string {
	if gameMap == nil {
		return ""
	}
	return fmt.Sprintf("{nodes: %v, links: %v, exits: %v}", gameMap.nodes, gameMap.links, gameMap.exits)
}

func (gameMap *GameMap) SetBobnetAgentIndex(index int) {
	if gameMap == nil {
		return
	}
	nbNodes := len(gameMap.nodes)
	if gameMap.bobnetAgentIndex >= 0 && gameMap.bobnetAgentIndex < nbNodes {
		gameMap.nodes[gameMap.bobnetAgentIndex].isBobnetAgentPresent = false
	}
	if index >= 0 && index < nbNodes {
		gameMap.bobnetAgentIndex = index
		gameMap.nodes[index].isBobnetAgentPresent = true
	} else {
		gameMap.bobnetAgentIndex = -1
	}
}

func (gameMap *GameMap) removeLink(link *Link) {
	if gameMap == nil || link == nil {
		return
	}
	link.node1.removeLink(link.node2)
	link.node2.removeLink(link.node1)
	linkIndex := -1
	for i, l := range gameMap.links {
		if (l.node1.index == link.node1.index || l.node1.index == link.node2.index) &&
			(l.node2.index == link.node1.index || l.node2.index == link.node2.index) {
			linkIndex = i
			break
		}
	}
	if linkIndex == -1 {
		return
	}
	if len(gameMap.links) == 1 {
		gameMap.links = []*Link{}
		return
	}
	tmp := gameMap.links[0]
	gameMap.links[0] = gameMap.links[linkIndex]
	gameMap.links[linkIndex] = tmp
	gameMap.links = gameMap.links[1:]
}

func buildMap() *GameMap {
	var nbNodes, nbLinks, nbExits int
	fmt.Scan(&nbNodes, &nbLinks, &nbExits)
	debug("nb nodes:", nbNodes)
	debug("nb links:", nbLinks)
	debug("nb exits:", nbExits)
	nodes := make([]*Node, nbNodes)
	for i := 0; i < nbNodes; i++ {
		nodes[i] = &Node{index: i, isBobnetAgentPresent: false, isExit: false}
	}
	links := make([]*Link, nbLinks)
	for i := 0; i < nbLinks; i++ {
		var nodeIndex1, nodeIndex2 int
		fmt.Scan(&nodeIndex1, &nodeIndex2)
		node1 := nodes[nodeIndex1]
		node2 := nodes[nodeIndex2]
		links[i] = &Link{node1: node1, node2: node2}
		node1.links = append(node1.links, node2)
		node2.links = append(node2.links, node1)
	}
	exits := make([]*Node, nbExits)
	for i := 0; i < nbExits; i++ {
		var exitNodeIndex int
		fmt.Scan(&exitNodeIndex)
		exits[i] = nodes[exitNodeIndex]
		exits[i].isExit = true
	}
	return &GameMap{bobnetAgentIndex: -1, exits: exits, links: links, nodes: nodes}
}

/*** SHORTEST PATH ***/

func getShortestPath(gameMap *GameMap, startIndex int, endIndex int) []int {
	// debug("Get shortest path between", startIndex, "and", endIndex)
	if startIndex == endIndex {
		return []int{}
	}
	nbNodes := len(gameMap.nodes)
	parentIndex := make([]int, nbNodes)
	for i, _ := range parentIndex {
		parentIndex[i] = -1
	}
	var queue []int
	queue = append(queue, startIndex)
	for {
		currentIndex := queue[0]
		// debug("current index", currentIndex)
		queue = queue[1:]
		currentNode := gameMap.nodes[currentIndex]
		for _, linkedNode := range currentNode.links {
			if linkedNode.index != startIndex && parentIndex[linkedNode.index] == -1 {
				parentIndex[linkedNode.index] = currentIndex
				queue = append(queue, linkedNode.index)
			}
		}
		if len(queue) == 0 || parentIndex[endIndex] != -1 {
			break
		}
	}
	// debug("Parents", parentIndex)
	if parentIndex[endIndex] == -1 {
		return []int{}
	}
	// debug("Create path")
	var path []int
	currentParentIndex := endIndex
	for parentIndex[currentParentIndex] != -1 {
		// debug("prepend", currentParentIndex)
		path = append([]int{currentParentIndex}, path...)
		currentParentIndex = parentIndex[currentParentIndex]
	}
	path = append([]int{startIndex}, path...)
	return path
}

func getBobnetPathToExit(channel chan []int, gameMap *GameMap, exitIndex int) {
	channel <- getShortestPath(gameMap, gameMap.bobnetAgentIndex, exitIndex)
}

func getBobnetPath(gameMap *GameMap) ([]int, error) {
	if gameMap == nil {
		return nil, errors.New("game map is missing")
	}
	nbExits := len(gameMap.exits)
	pathChannel := make(chan []int, nbExits)
	debug(nbExits, "paths to compute")
	for i := 0; i < nbExits; i++ {
		go getBobnetPathToExit(pathChannel, gameMap, gameMap.exits[i].index)
	}
	var path []int
	for i := 0; i < nbExits; i++ {
		pathToExit := <-pathChannel
		pathToExitLength := len(pathToExit)
		if pathToExitLength == 0 {
			continue
		}
		if path == nil || pathToExitLength < len(path) {
			path = pathToExit
		}
	}
	return path, nil
}

/*** LINK TO CUT ***/

func getLinkToCutFromPath(links []*Link, path []int) (*Link, error) {
	if path == nil || len(path) < 2 {
		return nil, errors.New("cannot get a link to cut, the path has less than 2 indexes")
	}
	nbLinks := len(links)
	currentPath := path
	pathLenght := len(currentPath)
	for len(currentPath) > 1 {
		index1 := currentPath[pathLenght-1]
		index2 := currentPath[pathLenght-2]
		for i := 0; i < nbLinks; i++ {
			link := links[i]
			if (link.node1.index == index1 || link.node1.index == index2) &&
				(link.node2.index == index1 || link.node2.index == index2) {
				return link, nil
			}
		}
		currentPath = currentPath[:pathLenght-2]
		pathLenght = len(currentPath)
	}
	return nil, errors.New("cannot get a link to cut from the path")
}

func getLinkToCutFromNode(node *Node) (*Link, error) {
	if node == nil || len(node.links) == 0 {
		return nil, errors.New("cannot get a link to cut, the node has no links")
	}
	return &Link{node1: node, node2: node.links[0]}, nil
}

/*** OUTPUTS ***/

func cutLink(gameMap *GameMap, link *Link) {
	if gameMap == nil || link == nil {
		return
	}
	gameMap.removeLink(link)
	fmt.Println(fmt.Sprintf("%d %d", link.node1.index, link.node2.index))
}

/*** MAIN ***/

func main() {
	start := time.Now()
	gameMap := buildMap()
	round := 0
	debug("Game map:", gameMap)
	debug("Init time:", time.Since(start))
	for {
		start = time.Now()
		round++
		var bobnetAgentIndex int
		fmt.Scan(&bobnetAgentIndex)
		gameMap.SetBobnetAgentIndex(bobnetAgentIndex)
		debug("Round", round, ":", gameMap.nodes)
		bobnetPath, bobnetPathError := getBobnetPath(gameMap)
		if bobnetPathError != nil {
			debug("Error getting bobnet shortest path:", bobnetPathError)
			continue
		}
		debug("Bobnet path", bobnetPath)
		linkToCut, linkToCutError := getLinkToCutFromPath(gameMap.links, bobnetPath)
		if linkToCutError != nil {
			debug("Error getting link to cut:", linkToCutError)
			debug("Get a link from bobnet agent node")
			linkToCut, linkToCutError = getLinkToCutFromNode(gameMap.nodes[gameMap.bobnetAgentIndex])
			if linkToCutError != nil {
				debug("Error getting link to cut:", linkToCutError)
				continue
			}
		}
		cutLink(gameMap, linkToCut)
		debug("Round time:", time.Since(start))
	}
}
