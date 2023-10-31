# Device Management Service Tx Validation Test

To run the tx validation test you can do the following:
`nix-shell --run "run-tx-validation-test"`

This starts a DMS, and sends an invalid job with an invalid Tx-Hash to it, then confirms the response
with the DMS that it ran the job and the job was successful.
