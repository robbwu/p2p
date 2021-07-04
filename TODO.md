# TODO

## Generate Paillier modulus as product of _safe_ primes

Up until now, the primes `p`,`q` generated for Paillier are only guaranteed to be such that `p,q = 3 mod 4`.
The original description of the protocol requires them to be **safe** primes. 
For efficiency reasons, we have not yet implemented this.

## Elliptic curve backend

Our implementation currently uses the `secp256k1` curve implementation from [decred/dcrd]("https://github.com/decred/dcrd/dcrec/secp256k1/v3").
We are working on a better interface that would seamlessly support multiple curve types.
One possibility we are exploring is the use of generics.

## Constant-time big.Int

Some Paillier operations may leak information about secrets used. 
Therefore, we are looking at using [cronokirby/safenum](https://github.com/cronokirby/safenum) for encryption and decryption of secret values

## Identifiable aborts

In some instances, it may be possible for the user of the library to guarantee a reliable broadcast channel (trusted third party in star topology for example).
It then makes sense to offer the option of identifying misbehaving parties, and replace the implicit echo broadcasts.

## lib-p2p examples

An example setup could use `libp2p` as a way of coordinating messages between parties .

## Specialized 2-2 protocol

The protocols currently work fine in a 2-2 setting,
but some specialization could be applied to make it more efficient
(for example, we would not need to use the echo broadcast).