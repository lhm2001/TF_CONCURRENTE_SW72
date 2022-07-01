package pe.edu.upc.services.impls;

import com.google.gson.Gson;
import com.google.gson.JsonObject;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import pe.edu.upc.entities.Transaction;
import pe.edu.upc.exception.ResourceNotFoundException;
import pe.edu.upc.repositories.TransactionRepository;
import pe.edu.upc.services.TransactionService;

import java.io.IOException;
import java.io.ObjectInputStream;
import java.io.ObjectOutputStream;
import java.io.Serializable;
import java.net.Socket;
import java.util.List;

@Service
public class TransactionServiceImpl implements TransactionService, Serializable {

    @Autowired
    private TransactionRepository transactionRepository;

    @Override
    public List<Transaction> findAll() throws Exception {
        return transactionRepository.findAll();
    }

    @Override
    public Transaction findById(Long aLong) throws Exception {
        return transactionRepository.findById(aLong)
                .orElseThrow(()->new ResourceNotFoundException("Transaction","Id",aLong));
    }

    @Override
    public Transaction createTransaction(Transaction transaction) {

        return transactionRepository.save(transaction);


    }

    @Override
    public void sendOverTcp(Transaction transaction) {
        ObjectOutputStream oos = null;

        Socket s = null;

        try {
            s = new Socket("127.0.0.1", 8000);
            oos = new ObjectOutputStream(s.getOutputStream());

            String id=transaction.getId().toString();
            String senderIr=transaction.getSenderId().toString();
            String receiverID=transaction.getReceiverId().toString();
            String amount=transaction.getAmount().toString();
            String json = "{\"cmd\": \"cliRegister\", \"sender\": \"faux\", \"data\": [\""+id+"\", \""+senderIr+"\", \""+receiverID+"\", \""+amount+"\"]}";
            oos.writeObject(json);


        } catch (Exception e) {
            e.printStackTrace();
        } finally {
            if (oos != null) {
                try {
                    oos.close();
                } catch (IOException e) {
                    throw new RuntimeException(e);
                }
            }
            if (s != null) {
                try {
                    s.close();
                } catch (IOException e) {
                    throw new RuntimeException(e);
                }
                System.out.println("Conexión cerrada!");
            }
        }
    }
}
