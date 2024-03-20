Feature: consult/1
  This feature is to test the consult/1 predicate.

  @great_for_documentation
  Scenario: Consult a Prolog program
  This scenario demonstrates how to consult (load) a Prolog program from a CosmWasm smart contract.

  Assuming the existence of a CosmWasm smart contract configured to store a Prolog program, we construct a URI to specifically
  identify this smart contract and pinpoint the Prolog program we want to consult via a query message.

    Given the CosmWasm smart contract "okp415ekvz3qdter33mdnk98v8whv5qdr53yusksnfgc08xd26fpdn3ts8gddht" and the behavior:
      """ yaml
      message: |
        {
          "object_data": {
            "id": "4cbe36399aabfcc7158ee7a66cbfffa525bb0ceab33d1ff2cff08759fe0a9b05"
          }
        }
      response: |
        hello("World!").
      """
    Given the program:
      """ prolog
      :-
        uri_encoded(query_value, '{"object_data":{"id": "4cbe36399aabfcc7158ee7a66cbfffa525bb0ceab33d1ff2cff08759fe0a9b05"}}', Query),
        atom_concat('cosmwasm:storage:okp415ekvz3qdter33mdnk98v8whv5qdr53yusksnfgc08xd26fpdn3ts8gddht?base64Decode=false&query=', Query, URI),
        consult(URI).
      """
    Given the query:
      """ prolog
      hello(Who).
      """
    When the query is run
    Then the answer we get is:
      """ yaml
      has_more: false
      variables: ["Who"]
      results:
      - substitutions:
        - variable: Who
          expression: "['W',o,r,l,d,!]"
      """


  @great_for_documentation
  Scenario: Consult a Prolog program which also consults another Prolog program
  This scenario demonstrates the capability of a Prolog program to consult another Prolog program. This is useful for
  modularizing Prolog programs and reusing code.

    Given the CosmWasm smart contract "okp415ekvz3qdter33mdnk98v8whv5qdr53yusksnfgc08xd26fpdn3ts8gddht" and the behavior:
      """ yaml
      message: |
        {
          "object_data": {
            "id": "4cbe36399aabfcc7158ee7a66cbfffa525bb0ceab33d1ff2cff08759fe0a9b05"
          }
        }
      response: |
        :- consult('cosmwasm:storage:okp412ssv28mzr02jffvy4x39akrpky9ykfafzyjzmvgsqqdw78yjevpqgmqnmk?query=%7B%22object_data%22%3A%7B%22id%22%3A%20%225d3933430d0a12794fae719e0db87b6ec5f549b2%22%7D%7D&base64Decode=false').
        a.
      """
    Given the CosmWasm smart contract "okp412ssv28mzr02jffvy4x39akrpky9ykfafzyjzmvgsqqdw78yjevpqgmqnmk" and the behavior:
      """ yaml
      message: |
        {
          "object_data": {
            "id": "5d3933430d0a12794fae719e0db87b6ec5f549b2"
          }
        }
      response: |
        b.
      """
    Given the query:
      """ prolog
      consult('cosmwasm:storage:okp415ekvz3qdter33mdnk98v8whv5qdr53yusksnfgc08xd26fpdn3ts8gddht?query=%7B%22object_data%22%3A%7B%22id%22%3A%20%224cbe36399aabfcc7158ee7a66cbfffa525bb0ceab33d1ff2cff08759fe0a9b05%22%7D%7D&base64Decode=false'),
      a, b.
      """
    When the query is run (limited to 2 solutions)
    Then the answer we get is:
      """ yaml
      has_more: false
      variables:
      results:
      - substitutions:
      """

  @great_for_documentation
  Scenario: Consult several Prolog programs
  This scenario demonstrates the consultation of several Prolog programs from different CosmWasm smart contracts.

    Given the CosmWasm smart contract "okp415ekvz3qdter33mdnk98v8whv5qdr53yusksnfgc08xd26fpdn3ts8gddht" and the behavior:
      """ yaml
      message: |
        {
          "object_data": {
            "id": "4cbe36399aabfcc7158ee7a66cbfffa525bb0ceab33d1ff2cff08759fe0a9b05"
          }
        }
      response: |
        program(a).
      """
    Given the CosmWasm smart contract "okp412ssv28mzr02jffvy4x39akrpky9ykfafzyjzmvgsqqdw78yjevpqgmqnmk" and the behavior:
      """ yaml
      message: |
        {
          "object_data": {
            "id": "5d3933430d0a12794fae719e0db87b6ec5f549b2"
          }
        }
      response: |
        program(b).
      """
    Given the program:
      """ prolog
        :- consult([
          'cosmwasm:storage:okp415ekvz3qdter33mdnk98v8whv5qdr53yusksnfgc08xd26fpdn3ts8gddht?query=%7B%22object_data%22%3A%7B%22id%22%3A%20%224cbe36399aabfcc7158ee7a66cbfffa525bb0ceab33d1ff2cff08759fe0a9b05%22%7D%7D&base64Decode=false',
          'cosmwasm:storage:okp412ssv28mzr02jffvy4x39akrpky9ykfafzyjzmvgsqqdw78yjevpqgmqnmk?query=%7B%22object_data%22%3A%7B%22id%22%3A%20%225d3933430d0a12794fae719e0db87b6ec5f549b2%22%7D%7D&base64Decode=false'
         ]).
      """
    Given the query:
      """ prolog
      source_file(File).
      """
    When the query is run (limited to 2 solutions)
    Then the answer we get is:
      """ yaml
      has_more: false
      variables: ["File"]
      results:
      - substitutions:
        - variable: File
          expression: "'cosmwasm:storage:okp412ssv28mzr02jffvy4x39akrpky9ykfafzyjzmvgsqqdw78yjevpqgmqnmk?query=%7B%22object_data%22%3A%7B%22id%22%3A%20%225d3933430d0a12794fae719e0db87b6ec5f549b2%22%7D%7D&base64Decode=false'"
      - substitutions:
        - variable: File
          expression: "'cosmwasm:storage:okp415ekvz3qdter33mdnk98v8whv5qdr53yusksnfgc08xd26fpdn3ts8gddht?query=%7B%22object_data%22%3A%7B%22id%22%3A%20%224cbe36399aabfcc7158ee7a66cbfffa525bb0ceab33d1ff2cff08759fe0a9b05%22%7D%7D&base64Decode=false'"
      """