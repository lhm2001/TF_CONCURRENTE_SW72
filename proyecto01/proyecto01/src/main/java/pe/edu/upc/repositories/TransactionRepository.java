package pe.edu.upc.repositories;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import pe.edu.upc.entities.Transaction;

@Repository
public interface TransactionRepository extends JpaRepository<Transaction,Long> {

}
