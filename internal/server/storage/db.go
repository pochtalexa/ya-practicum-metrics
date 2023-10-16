package storage

import (
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pochtalexa/ya-practicum-metrics/internal/server/flags"
	"github.com/rs/zerolog/log"
	"time"
)

type DBstore struct {
	DBconn *sql.DB
	Store  Store
}

func NewDBStore() *DBstore {
	return &DBstore{
		Store: Store{
			Gauges:   make(map[string]Gauge),
			Counters: make(map[string]Counter),
		},
	}
}

func InitConnDB() (*sql.DB, error) {
	ps := flags.FlagDBConn

	db, err := sql.Open("pgx", ps)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func PingDB(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return err
	}

	return nil
}

func InitialazeDB(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// создаем таблицу -gauge- для Gauge float64 - double precision
	gaugeTable := `CREATE TABLE IF NOT EXISTS gauge (
    								id		 serial PRIMARY KEY,
    								mname    varchar(40) UNIQUE,
    								val		 double precision	
                   )`
	// создаем таблицу -counter- для Counter int64 - integer
	counterTable := `CREATE TABLE IF NOT EXISTS counter (
    								id		 serial PRIMARY KEY,
    								mname    varchar(40) UNIQUE,
    								val		 integer	
                   )`

	if _, err := db.ExecContext(ctx, gaugeTable); err != nil {
		log.Err(err)
		return err
	}

	if _, err := db.ExecContext(ctx, counterTable); err != nil {
		log.Err(err)
		return err
	}

	return nil
}

func StoreMetricsToDB(d *DBstore) error {

	for k, v := range d.Store.Gauges {
		d.SetGauge(k, v)
	}

	for k, v := range d.Store.Counters {
		d.UpdateCounter(k, v)
	}

	return nil
}

func selectAllGauges(db *sql.DB) (map[string]Gauge, error) {
	var (
		result = make(map[string]Gauge)
		err    error
	)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, "SELECT mname, val from gauge")
	if err != nil {
		log.Err(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			mname string
			val   float64
		)

		err = rows.Scan(&mname, &val)
		if err != nil {
			log.Err(err)
			return nil, err
		}

		result[mname] = Gauge(val)
	}

	err = rows.Err()
	if err != nil {
		log.Err(err)
		return nil, err
	}

	return result, nil
}

func selectAllCounters(db *sql.DB) (map[string]Counter, error) {
	var (
		result = make(map[string]Counter)
		err    error
	)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, "SELECT mname, val from counter")
	if err != nil {
		log.Err(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			mname string
			val   int64
		)

		err = rows.Scan(&mname, &val)
		if err != nil {
			log.Err(err)
			return nil, err
		}

		result[mname] = Counter(val)
	}

	err = rows.Err()
	if err != nil {
		log.Err(err)
		return nil, err
	}

	return result, nil
}

func (d *DBstore) GetAllMetrics() Store {
	var (
		err error
	)

	d.Store.Gauges, err = selectAllGauges(d.DBconn)
	if err != nil {
		log.Err(err)
		panic(err)
	}

	d.Store.Counters, err = selectAllCounters(d.DBconn)
	if err != nil {
		log.Err(err)
		panic(err)
	}

	return d.Store
}

func (d *DBstore) GetGauge(name string) (Gauge, bool) {
	var result float64

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	row := d.DBconn.QueryRowContext(ctx, "SELECT val from gauge where mname = $1", name)

	err := row.Scan(&result)
	if err != nil {
		return -1, false
	}

	return Gauge(result), true
}

func (d *DBstore) GetGauges() map[string]Gauge {
	var (
		result = make(map[string]Gauge)
		err    error
	)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	rows, err := d.DBconn.QueryContext(ctx, "SELECT mname, val from gauge")
	if err != nil {
		log.Err(err)
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			mname string
			val   float64
		)

		err = rows.Scan(&mname, &val)
		if err != nil {
			log.Err(err)
			panic(err)
		}

		result[mname] = Gauge(val)
	}

	err = rows.Err()
	if err != nil {
		log.Err(err)
		panic(err)
	}

	return result
}

func (d *DBstore) SetGauge(name string, value Gauge) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	insertUpdate := `INSERT INTO gauge (mname, val) VALUES ($1, $2)
						ON CONFLICT (mname)
						DO UPDATE SET val = $3
						WHERE gauge.mname = $4`

	if _, err := d.DBconn.ExecContext(ctx, insertUpdate, name, value, value, name); err != nil {
		log.Err(err)
		panic(err)
	}
}

func (d *DBstore) GetCounter(name string) (Counter, bool) {
	var result int64

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	row := d.DBconn.QueryRowContext(ctx, "SELECT val from counter where mname = $1", name)

	err := row.Scan(&result)
	if err != nil {
		return -1, false
	}

	return Counter(result), true
}

func (d *DBstore) GetCounters() map[string]Counter {
	var (
		result = make(map[string]Counter)
		err    error
	)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	rows, err := d.DBconn.QueryContext(ctx, "SELECT mname, val from counter")
	if err != nil {
		log.Err(err)
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			mname string
			val   float64
		)

		err = rows.Scan(&mname, &val)
		if err != nil {
			log.Err(err)
			panic(err)
		}

		result[mname] = Counter(val)
	}

	err = rows.Err()
	if err != nil {
		log.Err(err)
		panic(err)
	}

	return result
}

func (d *DBstore) UpdateCounter(name string, value Counter) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	curVal, ok := d.GetCounter(name)
	if !ok {
		curVal = 0
	} else {
		curVal = curVal + value
	}

	insertUpdate := `INSERT INTO counter (mname, val) VALUES ($1, $2)
						ON CONFLICT (mname)
						DO UPDATE SET val = $3
						WHERE counter.mname = $4`

	if _, err := d.DBconn.ExecContext(ctx, insertUpdate, name, curVal, curVal, name); err != nil {
		log.Err(err)
		panic(err)
	}
}

func (d *DBstore) RestoreMetricsFromDB() error {
	var (
		err error
	)

	d.Store.Gauges, err = selectAllGauges(d.DBconn)
	if err != nil {
		log.Err(err)
		return err
	}

	d.Store.Counters, err = selectAllCounters(d.DBconn)
	if err != nil {
		log.Err(err)
		return err
	}

	return nil
}
