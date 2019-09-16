using Go = import "/go.capnp";
@0x92bf88982dabc9c9;
$Go.package("roompupdate");
$Go.import("github.com/jhaip/roomupdate");

struct RoomUpdates {
  updates @0 :List(RoomUpdate);

  struct RoomUpdate {
    type @0 :UpdateType;
    source @1 :Text;
    subscriptionId @2 :Text;  # used by PING and SUBSCRIPTION
    facts @3 :List(Fact);

    enum UpdateType {
        ping @0;
        claim @1;
        retract @2;
        subscribe @3;
        death @4;
        subscriptionDeath @5;
    }

    struct Fact {
        type @0 :Text;
        value @1 :Data;
    }
  }
}
