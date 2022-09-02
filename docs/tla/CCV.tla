--------------------------- MODULE CCV ---------------------------
(*
 * Modeling voting power relay between provider- and consumer chains in ICS.
 *
 * Simplifications:
 *   - We only track voting power, not bonded tokens
 *   - CCV channel creation is atomic and never fails/times out.
 *   - No new consumers join midway.
 *   - Block height is not modeled.
 * 
 * Jure Kukovec, 2022
 *)

EXTENDS Integers, Sequences, Apalache, typedefs

CONSTANT 
  \* The set of all nodes, which may take on a validator role.
  \* node \in Nodes is a validator <=> node \in DOMAIN votingPowerRunning
  \* @type: Set($node);
  Nodes, 
  \* The set of all consumer chains. Consumers may be removed 
  \* during execution, but not added.
  \* @type: Set($chain);
  ConsumerChains, 
  \* Time that needs to elapse, before a received VPC is considered
  \* mature on a chain.
  \* @type: $time;
  MaturityDelay,
  \* Time that needs to elapse, before a message is considered to have
  \* timed out (resulting in the removal of the related consumer chain).
  \* @type: $time;
  Timeout,
  \* Maximal time by which clocks are assumed to differ from the provider chain.
  \* Since consumer chains don't communicate, we don't care about 
  \* drift between tow consumers (though it's implicitly less than MaxDrift, if
  \* each differs from the provider chain by at most MaxDrift).
  \* The specification doesn't force clocks to maintain bounded drift, 
  \* but the invariants are only verified in cases where clocks never drift too far.
  \* @type: $time;
  MaxDrift

\* Provider chain only
VARIABLES
  \* Snapshots of the voting power on the provider chain, at the times
  \* when a VPC packet was sent.
  \* t \in DOMAIN votingPowerHist <=> VPC packet sent at time t
  \* @type: $time -> $votingPowerOnChain;
  votingPowerHist,
  \* Current voting power on the provider chain. 
  \* @type: $votingPowerOnChain;
  votingPowerRunning,
  \* Current set of active consumers. Inactive (by timeout) are dropped after
  \* each transition. May also drop arbitrarily.
  \* @type: Set($chain);
  activeConsumers,
  \* The set of VPC packet acknowledgements sent by consumer chains to the
  \* provider chain.
  \* @type: Set($ack);
  acks

\* Consumer chains or both
VARIABLES 
  \* Representation of the current voting power, as understood by consumer chains.
  \* Because consumer chains may not arbitrarily modify their own voting power,
  \* but must instead update in accordance to VPC packets received from the
  \* provider, it is sufficient to only track the last received packet.
  \* The voting power on chain c is then equal to votingPowerHist[votingPowerReferences[c]].
  \* @type: $chain -> $time;
  votingPowerReferences,
  \* The queues of VPC packets, waiting to be received by consumer chains.
  \* Note that a packet being placed in the channel is not considered 
  \* received by the consumer, until the receive-action is taken.
  \* @type: $chain -> Seq($packet);
  ccvChannelsPending,
  \* The queues of VPC packets, that have been received by consumer chains in the past.
  \* @type: $chain -> Seq($packet);
  ccvChannelsResolved,
  \* The current times of all chains (including the provider).
  \* @type: $chain -> $time;
  currentTimes,
  \* Bookkeeping of maturity times for received VPC packets.
  \* A consumer may only acknowledge a packet (i.e. notify the provider) after
  \* its local time exceeds the time designated in maturityTimes.
  \* For each consumer chain c, and VPC packet t sent by the provider,
  \* a) t \in DOMAIN maturityTimes[c] <=> c has received packet t 
  \* b) if t \in DOMAIN maturityTimes[c], then the acknowledge-action for t is 
  \*    guarded by currentTimes[c] >= maturityTimes[c][t]
  \* @type: $chain -> $packet -> $time;
  maturityTimes

\* Bookkeeping
VARIABLES 
  \* Name of last action, for debugging
  \* @type: Str;
  lastAction,
  \* VPC flag; Voting power may be considered to have changed, even if 
  \* the (TLA) value of votingPowerRunning does not (for example, due to a sequence
  \* of delegations and un-delegations, with a net 0 change in voting power).
  \* We use this flag to determine whether it is necessary to send a VPC packet.
  \* @type: Bool;
  votingPowerHasChanged,
  \* Invariant flag, TRUE iff clocks never drifted too much
  \* @type: Bool;
  boundedDrift

\* Helper tuples for UNCHANGED syntax
\* We don't track activeConsumers and lastAction in var tuples, because
\* they change each round.

providerVars ==
  << votingPowerHist, votingPowerRunning, acks >>

consumerVars ==
  << votingPowerReferences, ccvChannelsPending, ccvChannelsResolved, currentTimes, maturityTimes >>

\* @type: <<Bool, Bool>>;
bookkeepingVars == 
  << votingPowerHasChanged, boundedDrift >>


(*** NON-ACTION DEFINITIONS ***)

\* Some value not in Nat, for initialization
UndefinedTime == -1

\* Provider chain ID, assumed to be distinct from all consumer chain IDs
ProviderChain == "provider_OF_C"

\* Some value not in [Nodes -> Nat], for initialization
UndefinedPower == [node \in Nodes |-> -1]

\* All chains, including the provider. Used for the domain of shared
\* variables, e.g. currentTimes
Chains == ConsumerChains \union {ProviderChain}

\* According to https://github.com/cosmos/ibc/blob/main/spec/core/ics-004-channel-and-packet-semantics/README.md#receiving-packets
\* we need to use >=.
TimeoutGuard(a,b) == a >= b

\* @type: (Seq($packet), $time) => Bool;
TimeoutOnReception(channel, consumerT) ==
  /\ Len(channel) /= 0
  \* Head is always the oldest packet, so if there is a timeout for some packet, 
  \* there must be one for Head too
  /\ TimeoutGuard(consumerT, Head(channel) + Timeout)


\* @type: ($chain, $time, $packet -> $time) => Bool;
TimeoutOnAcknowledgement(c, providerT, maturity) ==
  \E packet \in DOMAIN maturity: 
    \* Note: Reception time = maturity[packet] - MaturityDelay
    /\ TimeoutGuard(providerT + MaturityDelay, maturity[packet] + Timeout)
    \* Not yet acknowledged
    /\ \A ack \in acks:
      \/ ack.chain /= c
      \/ ack.packetTime /= packet

\* Takes parameters, so primed and non-primed values can be passed
\* @type: ($chain, Seq($packet), $time, $time, $packet -> $time) => Bool;
PacketTimeoutForConsumer(c, channel, consumerT, providerT, maturity) == 
  \* Option 1: Timeout on reception
  \/ TimeoutOnReception(channel, consumerT)
  \* Option 2: Timeout on acknowledgement
  \/ TimeoutOnAcknowledgement(c, providerT, maturity)

\* Because we're not using functions with fixed domains, we can't use EXCEPT.
\* Thus, we need a helper method for domain-extension.
\* @type: (a -> b, a, b) => a -> b;
ExtendFnBy(f, k, v) ==
  [ 
    x \in DOMAIN f \union {k} |->
      IF x = k
      THEN v
      ELSE f[x]
  ]

\* Packets are set at unique times, monotonically increasing, the last
\* one is just the max in the votingPowerHist domain.
LastPacketTime == 
  LET Max2(a,b) == IF a >= b THEN a ELSE b IN
  ApaFoldSet(Max2, -1, DOMAIN votingPowerHist)

\* @type: ($chain, $time, $time) => $ack;
Ack(c, packetT, ackT) == 
  [chain |-> c, packetTime |-> packetT, ackTime |-> ackT]

\* @type: (Int, Int) => Int;
Delta(a,b) == IF a > b THEN a - b ELSE b - a

\* @type: (a -> Int, Set(a), Int) => Bool;
BoundedDeltas(fn, dom, bound) ==
  /\ dom \subseteq DOMAIN fn
  /\ \A v1, v2 \in dom:
    Delta(fn[v1], fn[v2]) <= bound

\* All the packets ever sent to c in the order they were sent in
\* @type: ($chain) => Seq($packet);
PacketOrder(c) == ccvChannelsResolved[c] \o ccvChannelsPending[c]

(*** ACTIONS ***)

Init == 
  /\ votingPowerHist = [t \in {} |-> UndefinedPower]
  /\ \E initValidators \in SUBSET Nodes:
    /\ initValidators /= {}
    /\ votingPowerRunning \in [initValidators -> Nat]
    /\ \A v \in initValidators: votingPowerRunning[v] > 0
  /\ activeConsumers = ConsumerChains
  /\ acks = {}
  /\ votingPowerReferences = [chain \in ConsumerChains |-> UndefinedTime]
  /\ ccvChannelsPending = [chain \in ConsumerChains |-> <<>>]
  /\ ccvChannelsResolved = [chain \in ConsumerChains |-> <<>>]
  /\ currentTimes = [c \in Chains |-> 0]
  /\ maturityTimes = [c \in ConsumerChains |-> [t \in {} |-> UndefinedTime]]
  /\ votingPowerHasChanged = FALSE
  /\ boundedDrift = TRUE
  /\ lastAction = "Init"

\* We combine all (un)delegate actions, as well as (un)bonding actions into an
\* abstract VotingPowerChange.
\* Since VPC packets are sent at most once at the end of each block,
\* the granularity wouldn't have added value to the model.
VotingPowerChange == 
  \E newValidators \in SUBSET Nodes:
    /\ newValidators /= {}
    /\ votingPowerRunning' \in [newValidators -> Nat]
    /\ \A v \in newValidators: votingPowerRunning'[v] > 0
    \* Note: votingPowerHasChanged' is set to true 
    \* even if votingPowerRunning' = votingPowerRunning
    /\ votingPowerHasChanged' = TRUE
    /\ UNCHANGED consumerVars
    /\ UNCHANGED << votingPowerHist, acks >>
    /\ lastAction' = "VotingPowerChange"

RcvPacket == 
  \E c \in activeConsumers:
    \* There must be a packet to be received
    /\ Len(ccvChannelsPending[c]) /= 0
    /\ LET packet == Head(ccvChannelsPending[c]) IN
      \* The voting power adjusts immediately, but the acknowledgement message 
      \* is sent later, on maturity 
      /\ votingPowerReferences' = [votingPowerReferences EXCEPT ![c] = packet]
      \* Maturity happens after MaturityDelay time has elapsed on c
      /\ maturityTimes' = [
        maturityTimes EXCEPT ![c] =
          ExtendFnBy(maturityTimes[c], packet, currentTimes[c] + MaturityDelay)
        ]
      /\ ccvChannelsResolved' = [ccvChannelsResolved EXCEPT ![c] = Append(@, packet)]
    \* Drop from channel, to unblock reception of other packets.
    /\ ccvChannelsPending' = [ccvChannelsPending EXCEPT ![c] = Tail(@)]
    /\ UNCHANGED providerVars
    /\ UNCHANGED currentTimes
    /\ UNCHANGED votingPowerHasChanged
    /\ lastAction' = "RcvPacket"

AckPacket == 
  \E c \in activeConsumers:
    \* Has been received
    \E packet \in DOMAIN maturityTimes[c]:
      \* Has matured 
      /\ currentTimes[c] >= maturityTimes[c][packet] 
      \* Hasn't been acknowledged
      /\ \A ack \in acks:
        \/ ack.chain /= c
        \/ ack.packetTime /= packet
      /\ acks' = acks \union { Ack(c, packet, currentTimes[c]) }
      /\ UNCHANGED consumerVars
      /\ UNCHANGED << votingPowerHist, votingPowerRunning >>
      /\ UNCHANGED votingPowerHasChanged
      /\ lastAction' = "AckPacket"

\* Partial action, always happens on Next
\* No need to purge data structures, we just don't access non-active indices
DropConsumers == 
  \E newActive \in SUBSET activeConsumers:
  /\ activeConsumers' = 
    { c \in newActive: ~PacketTimeoutForConsumer(c, ccvChannelsPending'[c], currentTimes'[c], currentTimes'[ProviderChain], maturityTimes'[c]) }
  
      
\* Partial action, always happens on EndBlock, may also happen independently
AdvanceTimeCore ==
  \E newTimes \in [Chains -> Nat]:
    \* None regress
    \* Does not guarantee strict time progression in AdvanceTime.
    \* In EndProviderBlockAndSendPacket, provider time is forced 
    \* to strictly progress with an additional constraint.
    /\ \A c \in Chains: newTimes[c] >= currentTimes[c]
    /\ currentTimes' = newTimes

\* Time may also elapse without EndProviderBlockAndSendPacket.
AdvanceTime ==
  /\ AdvanceTimeCore
  /\ UNCHANGED providerVars
  /\ UNCHANGED << votingPowerReferences, ccvChannelsPending, ccvChannelsResolved, maturityTimes >>
  /\ UNCHANGED votingPowerHasChanged
  /\ lastAction' = "AdvanceTime"

EndProviderBlockAndSendPacket ==
  \* Packets are only sent if there is a VPC
  /\ votingPowerHasChanged
  /\ ccvChannelsPending' = 
    [
      chain \in ConsumerChains |-> Append(
        ccvChannelsPending[chain], 
        \* a packet is just the current time, the VP can be read from votingPowerHist
        currentTimes[ProviderChain]
        )
    ]
  \* Reset flag for next block
  /\ votingPowerHasChanged' = FALSE 
  /\ votingPowerHist' = ExtendFnBy(votingPowerHist, currentTimes[ProviderChain], votingPowerRunning)
  \* packet sending forces time progression on provider
  /\ AdvanceTimeCore
  /\ currentTimes'[ProviderChain] > currentTimes[ProviderChain]
  /\ UNCHANGED <<votingPowerReferences, maturityTimes, ccvChannelsResolved>>
  /\ UNCHANGED <<votingPowerRunning, acks>>
  /\ lastAction' = "EndProviderBlockAndSendPacket"

\* If we ever drop all consumers, just do noting
Stagnate ==
  /\ UNCHANGED providerVars
  /\ UNCHANGED consumerVars
  /\ UNCHANGED activeConsumers
  /\ UNCHANGED bookkeepingVars
  /\ lastAction' = "Stagnate"

Next == 
  \//\ activeConsumers /= {}
    /\\/ EndProviderBlockAndSendPacket
      \/ VotingPowerChange
      \/ RcvPacket
      \/ AckPacket
      \/ AdvanceTime
    \* Drop timed out, maybe more
    /\ DropConsumers
    /\ boundedDrift' = boundedDrift /\
      BoundedDeltas(currentTimes', activeConsumers' \union {ProviderChain}, MaxDrift)
  \//\ activeConsumers = {}
    /\ Stagnate

(*** PROPERTIES/INVARIANTS ***)

\* VCS must also mature on provider
LastVCSMatureOnProvider ==
  LastPacketTime + MaturityDelay <= currentTimes[ProviderChain]

VPCUpdateInProgress == 
  \* some chain has pending packets
  \/ \E c \in activeConsumers: 
    \/ Len(ccvChannelsPending[c]) /= 0
    \/ \E packet \in DOMAIN maturityTimes[c]: maturityTimes[c][packet] < currentTimes[c]
  \* not enough time has elapsed on provider itself since last update
  \/ ~LastVCSMatureOnProvider

ActiveConsumersNotTimedOut ==
  \A c \in activeConsumers:
    ~PacketTimeoutForConsumer(c, ccvChannelsPending[c], currentTimes[c], currentTimes[ProviderChain], maturityTimes[c])

\* Sanity- predicates check that the data structures don't take on unexpected values
SanityVP == 
  /\ \A t \in DOMAIN votingPowerHist:
    LET VP == votingPowerHist[t] IN
    VP /= UndefinedPower <=> 
      \A node \in DOMAIN VP: VP[node] >= 0
  /\ \A node \in DOMAIN votingPowerRunning: votingPowerRunning[node] >= 0

SanityRefs ==
  \A c \in ConsumerChains:
    votingPowerReferences[c] < 0 <=> votingPowerReferences[c] = UndefinedTime

SanityMaturity ==
  \A c \in ConsumerChains:
    \A t \in DOMAIN maturityTimes[c]:
      LET mt == maturityTimes[c][t] IN
      mt < 0 <=> mt = UndefinedTime
  
Sanity ==
  /\ SanityVP
  /\ SanityRefs
  /\ SanityMaturity


\* Since the clocks may drift, any delay that exceeds
\* Timeout + MaxDrift is perceived as timeout on all chains
AdjustedTimeout == Timeout + MaxDrift

\* Any packet sent by the provider is either received within Timeout, or
\* the consumer chain is no longer considered active.
RcvdInTime ==
  \A t \in DOMAIN votingPowerHist:
    \A c \in activeConsumers:
      \* If c is still active after Timeout has elapsed from packet t broadcast ...
      TimeoutGuard(currentTimes[c], t + AdjustedTimeout) =>
        \* ... then c must have received packet t
        t \in DOMAIN maturityTimes[c]

\* Any packet received by the consumer and matured is acknowledged 
\* within Timeout of reception, or the consumer is no longer considered active.
AckdInTime ==
  \A c \in activeConsumers:
    \A t \in DOMAIN maturityTimes[c]:
      \* If c is still active after Timeout has elapsed from packet t reception ...
      \* Note: Reception time = maturityTimes[c][p] - MaturityDelay
      TimeoutGuard(currentTimes[ProviderChain] + MaturityDelay, maturityTimes[c][t] + AdjustedTimeout) =>
        \* ... then c must have acknowledged packet t
        \E ack \in acks:
          /\ ack.chain = c
          /\ ack.packetTime = t 


\* \* All packets are acked at the latest by Timeout, from all 
\* active consumers (or those consumers are removed from the active set)
\* It should be the case that RcvdInTime /\ AckdInTime => EventuallyAllAcks
EventuallyAllAcks == 
  \A t \in DOMAIN votingPowerHist:
    \* If a packet was sent at time t and enough time has elapsed, 
    \* s.t. all consumers should have responded ...
    TimeoutGuard(currentTimes[ProviderChain], t + 2 * AdjustedTimeout) =>
        \* then, all consumers have acked
        \A c \in activeConsumers:
          \E ack \in acks:
            /\ ack.chain = c
            /\ ack.packetTime = t


\* Invariants from https://github.com/cosmos/interchain-security/blob/main/docs/quality_assurance.md

(* 
4.10 - The provider chain's correctness is not affected by a consumer chain 
shutting down

What is "provider chain correctness"?
*)

(* 
4.11 - The provider chain can graciously handle a CCV packet timing out 
(without shutting down) - expected outcome: 
consumer chain shuts down and its state in provider CCV module is removed
*)
Inv411 == 
  boundedDrift =>
  \A c \in ConsumerChains:
    TimeoutOnReception(ccvChannelsPending[c], currentTimes[c]) =>
      c \notin activeConsumers

(* 
4.12 - The provider chain can graciously handle a StopConsumerChainProposal - 
expected outcome: consumer chain shuts down and its state 
in provider CCV module is removed.

What is "graciously handle"?
*)
  
(*
6.01 - Every validator set on any consumer chain MUST either be or have been 
a validator set on the provider chain.

In the current model, implicit through construction (votingPowerReferences)
*)
Inv601 ==
  \A c \in activeConsumers:
    LET ref == votingPowerReferences[c] IN
    ref /= UndefinedTime => ref \in DOMAIN votingPowerHist

(*
6.02 - Any update in the power of a validator val on the provider, as a result of
- (increase) Delegate() / Redelegate() to val
- (increase) val joining the provider validator set
- (decrease) Undelegate() / Redelegate() from val
- (decrease) Slash(val)
- (decrease) val leaving the provider validator set
MUST be present in a ValidatorSetChangePacket that is sent to all registered consumer chains
*)
Inv602 ==
  \A packet \in DOMAIN votingPowerHist:
    \A c \in activeConsumers:
      LET packetsToC == PacketOrder(c) IN
      \E i \in DOMAIN packetsToC:
        packetsToC[i] = packet

(*
6.03 - Every consumer chain receives the same sequence of 
ValidatorSetChangePackets in the same order.

Note: consider only prefixes on received packets (ccvChannelsResolved)
*)
Inv603 == 
  \A c1,c2 \in activeConsumers:
    PacketOrder(c1) = PacketOrder(c2)

(*
7.01 - For every ValidatorSetChangePacket received by a consumer chain at 
time t, a MaturedVSCPacket is sent back to the provider in the first block 
with a timestamp >= t + UnbondingPeriod

Modification: not necessarily _first_ block with that timestamp, 
since we don't model height _and_ time.
*)
Inv701 ==
  boundedDrift => AckdInTime

(*
7.02 - If an unbonding operation resulted in a ValidatorSetChangePacket sent
to all registered consumer chains, then it cannot complete before receiving
matching MaturedVSCPackets from these consumer chains 
(unless some of these consumer chains are removed)

We can define change completion, but we don't model it. Best approximation:
*)
Inv702 ==
  boundedDrift => EventuallyAllAcks

Inv ==
  \* /\ Sanity
  \* /\ ActiveConsumersNotTimedOut
  /\ (boundedDrift => 
      /\ RcvdInTime
      /\ AckdInTime
      )
  \* /\ RcvdInTime
  \* /\ AckdInTime
  


=============================================================================