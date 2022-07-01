package pe.edu.upc.services;

import pe.edu.upc.entities.Transaction;

public interface TransactionService extends CrudService<Transaction,Long>  {


    Transaction createTransaction(Transaction transaction);
    void sendOverTcp(Transaction transaction);

}
