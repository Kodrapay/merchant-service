package repositories

type MerchantRepository struct {
	dsn string
}

func NewMerchantRepository(dsn string) *MerchantRepository {
	return &MerchantRepository{
		dsn: dsn,
	}
}

// TODO: implement persistence layer for this microservice
