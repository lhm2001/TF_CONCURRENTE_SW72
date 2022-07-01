package pe.edu.upc.entities;

import lombok.Data;

import javax.persistence.*;

@Entity
@Table(name="transaction")
@Data
public class Transaction {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    private Long senderId;
    private Long receiverId;
    private Double amount;


}
