package repositories

import "log"

type MerchantRepository struct {
    dsn string
}

func NewMerchantRepository(dsn string) *MerchantRepository {
    log.Printf("MerchantRepository using DSN: %s", dsn)
    return &MerchantRepository{dsn: dsn}
}

// TODO: implement persistence for merchants, api keys, bank accounts.
