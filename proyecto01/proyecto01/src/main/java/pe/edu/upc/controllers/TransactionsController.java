package pe.edu.upc.controllers;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.validation.BindingResult;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.server.ResponseStatusException;
import pe.edu.upc.entities.Transaction;
import pe.edu.upc.services.TransactionService;

import javax.validation.Valid;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;


@Slf4j
@RestController
@RequestMapping("/transactions")
public class TransactionsController {

    @Autowired
    private TransactionService transactionService;


    @GetMapping(produces = MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<List<Transaction>> fetchAll() {
        try {
            List<Transaction> transactions = transactionService.findAll();
            return new ResponseEntity<List<Transaction>>(transactions, HttpStatus.OK);
        } catch (Exception e) {
            return new ResponseEntity<>(HttpStatus.INTERNAL_SERVER_ERROR);
        }
    }

    @GetMapping(path = "/{transactionId}", produces = MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<Transaction> fetchById(@PathVariable("transactionId") Long transactionId) {
        try {
            Transaction transaction = transactionService.findById(transactionId);
            if (transaction!=null) {
                return new ResponseEntity<Transaction>(transaction, HttpStatus.OK);
            } else {
                return new ResponseEntity<>(HttpStatus.NOT_FOUND);
            }
        } catch (Exception e) {
            return new ResponseEntity<>(HttpStatus.INTERNAL_SERVER_ERROR);
        }
    }

    @PostMapping
    public ResponseEntity<Transaction> createTransaction(@Valid @RequestBody Transaction transaction, BindingResult result){
        if (result.hasErrors()){
            throw new ResponseStatusException(HttpStatus.BAD_REQUEST,this.formatMessage(result));
        }
        Transaction transactionDB = transactionService.createTransaction(transaction);

        transactionService.sendOverTcp(transactionDB);

        return ResponseEntity.status(HttpStatus.CREATED).body(transactionDB);
    }


    private String formatMessage( BindingResult result){
        List<Map<String,String>> errors = result.getFieldErrors().stream()
                .map(err ->{
                    Map<String,String> error =  new HashMap<>();
                    error.put(err.getField(), err.getDefaultMessage());
                    return error;
                }).collect(Collectors.toList());
        ErrorMessage errorMessage = ErrorMessage.builder()
                .code("01")
                .messages(errors).build();
        ObjectMapper mapper = new ObjectMapper();
        String jsonString="";
        try {
            jsonString = mapper.writeValueAsString(errorMessage);
        } catch (JsonProcessingException e) {
            e.printStackTrace();
        }
        return jsonString;
    }

}
