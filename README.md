# go-ledger
worklod for POC purposes

# usecase
create a double ledger entry

# tables

    table transaction_type
    transaction_id|description  |created_at|
    --------------+-------------+----------+
    DEPOSIT       |DEPOSIT      |          |
    WITHDRAW      |WITHDRAW     |          |
    WIRE_TRANSFER |WIRE_TRANSFER|          |
    FINANCIAL_FEE |FINANCIAL_FEE|          |

    table ledger
    ledger_id           |fk_account_type_id|description           |created_at|
    --------------------+------------------+----------------------+----------+
    ACC-BANK:ASSET      |ASSET             |ACCOUNT-BANK:ASSET    |          |
    ACC-BANK:LIABILITY  |LIABILITY         |ACCOUNT-BANK:LIABILITY|          |
    INTER-BANK:ASSET    |ASSET             |INTER-BANK:ASSET      |          |
    INTER-BANK:LIABILITY|LIABILITY         |INTER-BANK:LIABILITY  |          |

    table transaction
    id  |currency|description  |transaction_at               |
    ----+--------+-------------+-----------------------------+
    1   |BRL     |DEPOSIT      |2025-04-07 21:28:43.680 -0300|
    2   |BRL     |DEPOSIT      |2025-04-08 20:52:10.106 -0300|
    1958|BRL     |WITHDRAW     |2025-05-02 22:25:32.787 -0300|
    4   |BRL     |WITHDRAW     |2025-04-08 23:45:48.835 -0300|

    table transaction_detail
    fk_transaction_id|id   |fk_ledger_id  |fk_account_id|debit_amount|credit_amount|created_at                  
    -----------------+-----+--------------+-------------+------------+-------------+----------------------------
                    1|    1|BANK:LIABILITY|ACC-1        |           0|        10000|2025-04-07 21:28:43.687 -030
                    1|    2|BANK:LIABILITY|ACC-BANK     |       10000|            0|2025-04-07 21:28:43.693 -030
                    2|    3|BANK:LIABILITY|ACC-1        |           0|        10000|2025-04-08 20:52:10.115 -030
                    2|    4|BANK:LIABILITY|ACC-BANK     |       10000|            0|2025-04-08 20:52:10.123 -030
                1328| 2653|BANK:LIABILITY|ACC-263      |        4100|            0|2025-05-02 22:20:46.791 -030
                1328| 2654|BANK:LIABILITY|ACC-BANK     |           0|         4100|2025-05-02 22:20:46.794 -030