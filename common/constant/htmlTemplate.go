package constant

const NodeStatusHTMLTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8" />
</head>
<body>
<div>
    YOUR ZOOBC NODE IS ONLINE
    <br/><br/>
    <table border="1">
        <tr>
            <td>Version</td>
            <td>{{.Version}}</td>
        </tr>
        <tr>
            <td>Last Block Height (Mainchain)</td>
            <td>{{.LastMainBlockHeight}}</td>
        </tr>
        <tr>
            <td>Last Block Hash (Mainchain)</td>
            <td>{{.LastMainBlockHash}}</td>
        </tr>
        <tr>
            <td>Last Block Height (Spinechain)</td>
            <td>{{.LastSpineBlockHeight}}</td>
        </tr>
        <tr>
            <td>Last Block Hash (Spinechain)</td>
            <td>{{.LastSpineBlockHash}}</td>
        </tr>
        <tr>
            <td>Node Public Key</td>
            <td>{{.NodePublicKey}}</td>
        </tr>
        <tr>
            <td>Resolved Peer</td>
            <td>{{.ResolvedPeers}}</td>
        </tr>
        <tr>
            <td>Unresolved Peer</td>
            <td>{{.UnresolvedPeers}}</td>
        </tr>
        <tr>
            <td>Smithing Index</td>
            <td>{{.BlocksmithIndex}}</td>
        </tr>
    </table>
</div>
</body>
</html>
`
