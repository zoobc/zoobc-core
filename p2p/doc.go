package p2p

/*
p2p package is a sub-module of zoobc core application, this package wrap all the functionality for p2p communication, including
 - spawn go routines that:
  - request for more peers
  - resolve peers
  - blacklist peers
 - open rpc endpoint
 - send request to another node p2p endpoint, and
 - managing peers (strategy package)

Another functionality that is bind to this package is listeners, since broadcasting information to another peer will be
done in this package, for block, transaction, and another data broadcast listener, will be registered here.
*/
