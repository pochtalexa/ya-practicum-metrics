package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pochtalexa/ya-practicum-metrics/internal/server/flags"
	"github.com/pochtalexa/ya-practicum-metrics/internal/server/models"
	"github.com/rs/zerolog/log"
	"github.com/sethvargo/go-retry"
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

	// не возвращает ошибку елси нет коннетка к БД
	db, err := sql.Open("pgx", ps)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func PingDB(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return err
	}

	return nil
}

func InitialazeDB(db *sql.DB) error {
	var err error
	b := retry.NewFibonacci(1 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
    								val		 bigint	
                   )`

	err = retry.Do(ctx, retry.WithMaxRetries(3, b), func(ctx context.Context) error {

		if _, err = db.Exec(gaugeTable); err != nil {
			log.Info().Err(err).Msg("DB gaugeTable error")
			return retry.RetryableError(err)
		}

		if _, err = db.Exec(counterTable); err != nil {
			log.Info().Err(err).Msg("DB counterTable error")
			return retry.RetryableError(err)
		}

		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func StoreMetricsToDB(d *DBstore) error {
	var err error

	for k, v := range d.Store.Gauges {
		err = d.SetGauge(k, v)
		if err != nil {
			return err
		}
	}

	for k, v := range d.Store.Counters {
		err = d.UpdateCounter(k, v)
		if err != nil {
			return err
		}
	}

	return nil
}

// отрабатывает с retry
func selectAllGauges(db *sql.DB) (map[string]Gauge, error) {
	var (
		result = make(map[string]Gauge)
		err    error
		rows   *sql.Rows
	)
	b := retry.NewFibonacci(1 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = retry.Do(ctx, retry.WithMaxRetries(3, b), func(ctx context.Context) error {
		rows, err = db.Query("SELECT mname, val from gauge")
		if err != nil {
			log.Info().Err(err).Msg("DB selectAllGauges QueryContext error")
			return retry.RetryableError(err)
		}
		defer rows.Close()

		for rows.Next() {
			var (
				mname string
				val   float64
			)

			err = rows.Scan(&mname, &val)
			if err != nil {
				log.Info().Err(err).Msg("DB selectAllGauges rows.Scan error")
				return retry.RetryableError(err)
			}

			result[mname] = Gauge(val)
		}

		err = rows.Err()
		if err != nil {
			log.Info().Err(err).Msg("DB selectAllGauges rows.Err error")
			return retry.RetryableError(err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// отрабатывает с retry
func selectAllCounters(db *sql.DB) (map[string]Counter, error) {
	var (
		result = make(map[string]Counter)
		err    error
		rows   *sql.Rows
	)
	b := retry.NewFibonacci(1 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = retry.Do(ctx, retry.WithMaxRetries(3, b), func(ctx context.Context) error {
		rows, err = db.QueryContext(ctx, "SELECT mname, val from counter")
		if err != nil {
			log.Info().Err(err).Msg("DB selectAllCounters QueryContext error")
			return retry.RetryableError(err)
		}
		defer rows.Close()

		for rows.Next() {
			var (
				mname string
				val   int64
			)

			err = rows.Scan(&mname, &val)
			if err != nil {
				log.Info().Err(err).Msg("DB selectAllCounters rows.Scan error")
				return retry.RetryableError(err)
			}

			result[mname] = Counter(val)
		}

		err = rows.Err()
		if err != nil {
			log.Info().Err(err).Msg("DB selectAllCounters rows.Err error")
			return retry.RetryableError(err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (d *DBstore) GetAllMetrics() (Store, error) {
	var err error

	d.Store.Gauges, err = selectAllGauges(d.DBconn)
	if err != nil {
		log.Info().Err(err).Msg("DB selectAllGauges error")
		return Store{}, err
	}

	d.Store.Counters, err = selectAllCounters(d.DBconn)
	if err != nil {
		log.Info().Err(err).Msg("DB selectAllCounters error")
		return Store{}, err

	}

	return d.Store, nil
}

func (d *DBstore) GetGauge(name string) (Gauge, bool, error) {
	var result float64
	var err error
	b := retry.NewFibonacci(1 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = retry.Do(ctx, retry.WithMaxRetries(3, b), func(ctx context.Context) error {
		err = d.DBconn.QueryRow("SELECT val from gauge where mname = $1", name).Scan(&result)
		if errors.Is(err, sql.ErrNoRows) {
			return retry.RetryableError(err)
		} else if err != nil {
			log.Info().Err(err).Msg("DB selectAllCounters error")
			return err
		}

		return nil
	})

	if err != nil {
		return Gauge(result), false, err
	}
	return Gauge(result), true, nil
}

func (d *DBstore) GetGauges() (map[string]Gauge, error) {
	var (
		result = make(map[string]Gauge)
		err    error
		rows   *sql.Rows
	)
	b := retry.NewFibonacci(1 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = retry.Do(ctx, retry.WithMaxRetries(3, b), func(ctx context.Context) error {
		rows, err = d.DBconn.Query("SELECT mname, val from gauge")
		if err != nil {
			log.Info().Err(err).Msg("DB QueryContext GetGauges error")
			return retry.RetryableError(err)
		}
		defer rows.Close()

		for rows.Next() {
			var (
				mname string
				val   float64
			)

			err = rows.Scan(&mname, &val)
			if err != nil {
				log.Info().Err(err).Msg("DB rows.Scan QueryContext GetGauges error")
				return retry.RetryableError(err)
			}

			result[mname] = Gauge(val)
		}

		err = rows.Err()
		if err != nil {
			log.Info().Err(err).Msg("DB rows.Err QueryContext GetGauges error")
			return retry.RetryableError(err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (d *DBstore) SetGauge(name string, value Gauge) error {
	var err error
	b := retry.NewFibonacci(1 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	insertUpdate := `INSERT INTO gauge (mname, val) VALUES ($1, $2)
						ON CONFLICT (mname)
						DO UPDATE SET val = $3
						WHERE gauge.mname = $4`

	err = retry.Do(ctx, retry.WithMaxRetries(3, b), func(ctx context.Context) error {

		_, err = d.DBconn.Exec(insertUpdate, name, value, value, name)
		if err != nil {
			log.Info().Err(err).Msg("DB ExecContext SetGauge error")
			return retry.RetryableError(err)
		}

		return nil
	})

	return err
}

// отрабатывает с retry
func (d *DBstore) GetCounter(name string) (Counter, bool, error) {
	var result int64
	var err error
	b := retry.NewFibonacci(1 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = retry.Do(ctx, retry.WithMaxRetries(3, b), func(ctx context.Context) error {

		err = d.DBconn.QueryRowContext(ctx, "SELECT val from counter where mname = $1", name).Scan(&result)
		if errors.Is(err, sql.ErrNoRows) {
			return err
		} else if err != nil {
			log.Info().Err(err).Msg("DB GetCounter QueryRowContext error")
			return retry.RetryableError(err)
		}

		return nil
	})

	if errors.Is(err, sql.ErrNoRows) {
		return Counter(result), false, nil
	}
	if err != nil {
		return Counter(result), false, err
	}
	return Counter(result), true, nil
}

func (d *DBstore) GetCounters() (map[string]Counter, error) {
	var (
		result = make(map[string]Counter)
		rows   *sql.Rows
		err    error
	)
	b := retry.NewFibonacci(1 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = retry.Do(ctx, retry.WithMaxRetries(3, b), func(ctx context.Context) error {
		rows, err = d.DBconn.QueryContext(ctx, "SELECT mname, val from counter")
		if err != nil {
			log.Info().Err(err).Msg("DB GetCounters QueryContext error")
			return retry.RetryableError(err)
		}
		defer rows.Close()

		for rows.Next() {
			var (
				mname string
				val   float64
			)

			err = rows.Scan(&mname, &val)
			if err != nil {
				log.Info().Err(err).Msg("DB rows.Scan QueryContext GetCounters error")
				return retry.RetryableError(err)
			}

			result[mname] = Counter(val)
		}

		err = rows.Err()
		if err != nil {
			log.Info().Err(err).Msg("DB rows.Err QueryContext GetCounters error")
			return retry.RetryableError(err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (d *DBstore) UpdateCounter(name string, value Counter) error {
	b := retry.NewFibonacci(1 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	curVal, ok, err := d.GetCounter(name)
	if err != nil {
		log.Info().Err(err).Msg("DB GetCounter UpdateCounter GetCounters error")
		return err
	}
	if !ok {
		curVal = 0
	}
	curVal = curVal + value

	insertUpdate := `INSERT INTO counter (mname, val) VALUES ($1, $2)
						ON CONFLICT (mname)
						DO UPDATE SET val = $3
						WHERE counter.mname = $4`

	err = retry.Do(ctx, retry.WithMaxRetries(3, b), func(ctx context.Context) error {

		_, err = d.DBconn.Exec(insertUpdate, name, curVal, curVal, name)
		if err != nil {
			log.Info().Err(err).Msg("DB UpdateCounter error")
			return retry.RetryableError(err)
		}

		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func (d *DBstore) RestoreMetricsFromDB() error {
	var err error

	// selectAllGauges возвращает retry.RetryableError(err)
	d.Store.Gauges, err = selectAllGauges(d.DBconn)
	if err != nil {
		log.Info().Err(err).Msg("DB selectAllGauges RestoreMetricsFromDB error")
		return err
	}

	// selectAllCounters возвращает retry.RetryableError(err)
	d.Store.Counters, err = selectAllCounters(d.DBconn)
	if err != nil {
		log.Info().Err(err).Msg("DB selectAllCounters RestoreMetricsFromDB error")
		return err
	}

	return nil
}

func (d *DBstore) UpdateMetricBatch(reqJSON []models.Metrics) error {
	var err error
	b := retry.NewFibonacci(1 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := d.DBconn.Begin()
	if err != nil {
		retry.RetryableError(err)
	}
	defer tx.Rollback()

	insertUpdateGauge := `INSERT INTO gauge (mname, val) VALUES ($1, $2)
						ON CONFLICT (mname)
						DO UPDATE SET val = $3
						WHERE gauge.mname = $4`

	err = retry.Do(ctx, retry.WithMaxRetries(3, b), func(ctx context.Context) error {

		for _, v := range reqJSON {
			if v.MType == "gauge" {
				if _, err = tx.ExecContext(ctx, insertUpdateGauge, v.ID, v.Value, v.Value, v.ID); err != nil {
					log.Info().Err(err).Msg("DB tx insertUpdateGauge error")
					return retry.RetryableError(err)
				}
			} else if v.MType == "counter" {
				if err = d.UpdateCounter(v.ID, Counter(*v.Delta)); err != nil {
					return retry.RetryableError(err)
				}
			} else {
				err := fmt.Errorf("can not get val for %v from reqJSON", v.ID)
				log.Info().Err(err).Msg("reqJSON error")
				return err
			}
		}

		err = tx.Commit()
		if err != nil {
			log.Info().Err(err).Msg("DB tx.Commit error")
			return retry.RetryableError(err)
		}

		return nil
	})

	if err != nil {
		return err
	}
	return nil
}
