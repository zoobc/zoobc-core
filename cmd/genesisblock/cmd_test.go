package genesisblock

// func Test_PubKeyFromSeed(t *testing.T) {
// 	// seed := "spree uplifted stapling quotable disfigure lair deduct untying timothy maggot cryptic unrigged"
// 	ed := crypto.NewEd25519Signature()
// 	seed := util.GetSecureRandomSeed()
// 	privKey := ed.GetPrivateKeyFromSeed(seed)
// 	pubKey, _ := ed.GetPublicKeyFromPrivateKey(privKey)
// 	id, _ := address.EncodeZbcID(constant.PrefixZoobcNodeAccount, pubKey)
// 	fmt.Printf("seed: %v\nid: %s\n", seed, id)
// 	// getPubKeyFromAddress := func(address string) ([]byte, error) {
// 	// 	publicKey, err := base64.URLEncoding.DecodeString(address)
// 	// 	if err != nil {
// 	//
// 	// 		return nil, blocker.NewBlocker(blocker.AppErr, err.Error())
// 	// 	}
// 	// 	// Needs to check the checksum bit at the end, and if valid,
// 	// 	if publicKey[32] != util.GetChecksumByte(publicKey[:32]) {
// 	// 		return nil, blocker.NewBlocker(blocker.AppErr, "address checksum failed")
// 	// 	}
// 	// 	return publicKey[:32], nil
// 	// }
// 	// pubKey2, _ := getPubKeyFromAddress("5Zup5YFOgVUZHRtoz4E9-8Ki2fN0U8DTHCM6WxbvgGg_")
// 	// if !reflect.DeepEqual(pubKey, pubKey2) {
// 	// 	t.Errorf("expect: %v\ngot: %v\n", pubKey, pubKey2)
// 	// }
// }
//
// func Test_generateGenesisFiles(t *testing.T) {
// 	t.Run("Convert-PreregisteredAccount", func(t *testing.T) {
// 		var (
// 			bcState, preRegisteredNodes []genesisEntry
// 			err                         error
// 		)
// 		file, err := ioutil.ReadFile(path.Join(getRootPath(), fmt.Sprintf("./genesisblock/templates/%s.preRegisteredNodes.json", "develop")))
// 		fmt.Printf("err: %v\n", err)
// 		if err == nil {
// 			err = json.Unmarshal(file, &preRegisteredNodes)
// 			if err != nil {
// 				log.Fatalf("preRegisteredNodes.json parsing error: %s", err)
// 			}
//
// 			// merge duplicates: if preRegisteredNodes contains entries that are in db too, add the parameters that are't available in db,
// 			// which is are NodeSeed and Smithing
// 			for _, prNode := range preRegisteredNodes {
// 				found := false
// 				for i, e := range bcState {
// 					if prNode.AccountAddress != e.AccountAddress {
// 						continue
// 					}
// 					bcState[i].NodeSeed = prNode.NodeSeed
// 					bcState[i].Smithing = prNode.Smithing
// 					pubKey, err := base64.StdEncoding.DecodeString(prNode.NodeAccountAddress)
// 					if err != nil {
// 						log.Fatal(err)
// 					}
// 					bcState[i].NodePublicKey = pubKey
// 					found = true
// 					break
// 				}
// 				if !found {
// 					prNode.NodePublicKey, err = base64.StdEncoding.DecodeString(prNode.NodeAccountAddress)
// 					if err != nil {
// 						log.Fatal(err)
// 					}
// 					bcState = append(bcState, prNode)
// 				}
// 			}
// 		}
//
// 		// file, err = ioutil.ReadFile(path.Join(getRootPath(), fmt.Sprintf("./genesisblock/templates/%s.genesisAccountAddresses.json", "alpha")))
// 		// var idx int
// 		// for idx = 0; idx < extraNodesCount; idx++ {
// 		// 	bcState = append(bcState, generateRandomGenesisEntry(idx, ""))
// 		// }
// 		// if err == nil {
// 		// 	// read custom addresses from file
// 		// 	err = json.Unmarshal(file, &preRegisteredAccountAddresses)
// 		// 	if err != nil {
// 		// 		log.Fatalf("preRegisteredAccountAddresses.json parsing error: %s", err)
// 		// 	}
// 		// 	if idx == 0 {
// 		// 		idx--
// 		// 	}
// 		// 	for _, preRegisteredAccountAddress := range preRegisteredAccountAddresses {
// 		// 		idx++
// 		// 		bcState = append(bcState, generateRandomGenesisEntry(idx, preRegisteredAccountAddress.AccountAddress))
// 		// 	}
// 		// }
// 		getPubKeyFromAddress := func(address string) ([]byte, error) {
// 			publicKey, err := base64.URLEncoding.DecodeString(address)
// 			if err != nil {
//
// 				return nil, blocker.NewBlocker(blocker.AppErr, err.Error())
// 			}
// 			// Needs to check the checksum bit at the end, and if valid,
// 			if publicKey[32] != util.GetChecksumByte(publicKey[:32]) {
// 				return nil, blocker.NewBlocker(blocker.AppErr, "address checksum failed")
// 			}
// 			return publicKey[:32], nil
// 		}
// 		for i, node := range bcState {
// 			fmt.Printf("i: %d\t", i)
// 			newNodeAddress, _ := address.EncodeZbcID(constant.PrefixZoobcNodeAccount, node.NodePublicKey)
// 			oldAddPubKey, _ := getPubKeyFromAddress(node.AccountAddress)
// 			newAddress, _ := address.EncodeZbcID(constant.PrefixZoobcNormalAccount, oldAddPubKey)
// 			fmt.Printf("old-node: %v\tnew-node: %v\n\told-account: %v\tnew-account: %v\n",
// 				node.NodeAccountAddress, newNodeAddress,
// 				node.AccountAddress, newAddress)
// 		}
// 	})
// }
//
// func Test_StringByteLength(t *testing.T) {
// 	var ed = crypto.NewEd25519Signature()
//
// 	for i := 0; i < 100000; i++ {
// 		nodeSeed := util.GetSecureRandomSeed()
// 		nodePrivateKey := ed.GetPrivateKeyFromSeed(nodeSeed)
// 		nodePublicKey := nodePrivateKey[32:]
// 		nodeAccountString, _ := address.EncodeZbcID(constant.PrefixZoobcNodeAccount, nodePublicKey)
// 		normalAccountString, _ := address.EncodeZbcID(constant.PrefixZoobcNormalAccount, nodePublicKey)
// 		if len(nodeAccountString) != len([]byte(nodeAccountString)) {
// 			fmt.Printf("i: %d\n", i)
// 			t.Errorf("node\tstring-length: %d\tbyte-length: %d\n", len(nodeAccountString), len([]byte(nodeAccountString)))
// 		}
// 		if len(normalAccountString) != len([]byte(normalAccountString)) {
// 			fmt.Printf("i: %d\n", i)
// 			t.Errorf("node\tstring-length: %d\tbyte-length: %d\n", len(normalAccountString), len([]byte(normalAccountString)))
// 		}
// 	}
// }
