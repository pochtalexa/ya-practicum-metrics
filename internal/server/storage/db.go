package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pochtalexa/ya-practicum-metrics/internal/server/flags"
	"github.com/pochtalexa/ya-practicum-metrics/internal/server/models"
	"github.com/rs/zerolog/log"
	"net"
	"strconv"
	"strings"
	"time"
)

type DBstore struct {
	DBconn *sql.DB
	Store  Store
}

type getGaugeCounter struct {
	counter int64
	gauge   float64
	name    string
}

func NewDBStore() *DBstore {
	return &DBstore{
		Store: Store{
			Gauges:   make(map[string]Gauge),
			Counters: make(map[string]Counter),
		},
	}
}

func txCommitRetry(tx *sql.Tx) error {
	var err error

	waiteIntervals := []int64{0, 1, 3, 5}
	waiteIntervalsLen := len(waiteIntervals) - 1
	for k, v := range waiteIntervals {
		var (
			pgErr  *pgconn.PgError
			netErr net.Error
		)

		time.Sleep(time.Second * time.Duration(v))
		err = tx.Commit()

		if (errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) || errors.As(err, &netErr)) &&
			k != waiteIntervalsLen {
			log.Info().
				Err(err).
				Str("attempt", strconv.FormatInt(int64(k), 10)).
				Msg("DB tx.Commit attempt error")
			continue
		} else if (errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) || errors.As(err, &netErr)) &&
			k == waiteIntervalsLen {
			log.Info().
				Err(err).
				Str("attempt", strconv.FormatInt(int64(k), 10)).
				Msg("DB tx.Commit attempt error")
			return err
		} else if err != nil {
			log.Info().Msg(err.Error())
			return err
		}
		break
	}

	log.Info().Msg("DB tx.Commit success")
	return nil
}

func QueryRowContextRetry(ctx context.Context, conn *sql.DB, reqDB string, params []string, result *getGaugeCounter) error {
	var (
		err error
		row *sql.Row
	)

	waiteIntervals := []int64{0, 1, 3, 5}
	waiteIntervalsLen := len(waiteIntervals) - 1
	for k, v := range waiteIntervals {
		var (
			pgErr  *pgconn.PgError
			netErr net.Error
		)

		time.Sleep(time.Second * time.Duration(v))
		row = conn.QueryRowContext(ctx, reqDB, strings.Join(params, ","))

		if result.name == "counter" {
			err = row.Scan(&result.counter)
		} else if result.name == "gauge" {
			err = row.Scan(&result.gauge)
		} else {
			panic(err)
		}

		if (errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) || errors.As(err, &netErr)) &&
			k != waiteIntervalsLen {
			log.Info().
				Err(err).
				Str("attempt", strconv.FormatInt(int64(k), 10)).
				Msg("DB QueryRowContext attempt error")
			continue
		} else if (errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) || errors.As(err, &netErr)) &&
			k == waiteIntervalsLen {
			log.Info().
				Err(err).
				Str("attempt", strconv.FormatInt(int64(k), 10)).
				Msg("DB QueryRowContext attempt error")
			return err
		} else if err != nil {
			log.Info().Msg(err.Error())
			return err
		}
		break
	}

	log.Info().Msg("DB QueryRowContextRetry success")
	return nil
}

func QueryContextRetry(ctx context.Context, db *sql.DB, reqDB string) (*sql.Rows, error) {
	var (
		err  error
		rows *sql.Rows
	)

	waiteIntervals := []int64{0, 1, 3, 5}
	waiteIntervalsLen := len(waiteIntervals) - 1
	for k, v := range waiteIntervals {
		var (
			pgErr  *pgconn.PgError
			netErr net.Error
		)

		time.Sleep(time.Second * time.Duration(v))
		rows, err = db.QueryContext(ctx, reqDB)
		if (errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) || errors.As(err, &netErr)) &&
			k != waiteIntervalsLen {
			log.Info().
				Err(err).
				Str("attempt", strconv.FormatInt(int64(k), 10)).
				Msg("DB QueryContextRetry attempt error")
			continue
		} else if (errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) || errors.As(err, &netErr)) &&
			k == waiteIntervalsLen {
			log.Info().
				Err(err).
				Str("attempt", strconv.FormatInt(int64(k), 10)).
				Msg("DB QueryContextRetry attempt error")
			return nil, err
		} else if err != nil {
			return nil, err
		}
		break
	}

	return rows, nil
}

func ExecContextRetry(ctx context.Context, db *sql.DB, reqDB string, params []string) error {
	var (
		err    error
		netErr net.Error
		pgErr  *pgconn.PgError
	)

	waiteIntervals := []int64{0, 1, 3, 5}
	waiteIntervalsLen := len(waiteIntervals) - 1
	for k, v := range waiteIntervals {
		time.Sleep(time.Second * time.Duration(v))
		if len(params) > 0 {
			_, err = db.ExecContext(ctx, reqDB, strings.Join(params, ","))
		} else {
			_, err = db.ExecContext(ctx, reqDB)
		}
		if (errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) || errors.As(err, &netErr)) &&
			k != waiteIntervalsLen {
			log.Info().
				Err(err).
				Str("attempt", strconv.FormatInt(int64(k), 10)).
				Msg("DB ExecContextRetry attempt error")
			continue
		} else if (errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) || errors.As(err, &netErr)) &&
			k == waiteIntervalsLen {
			log.Info().
				Err(err).
				Str("attempt", strconv.FormatInt(int64(k), 10)).
				Msg("DB ExecContextRetry attempt error")
			return err
		} else if err != nil {
			return err
		}
		break
	}

	return nil
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return err
	}

	return nil
}

func InitialazeDB(db *sql.DB) error {
	var err error

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
    								val		 integer	
                   )`

	err = ExecContextRetry(ctx, db, gaugeTable, []string{})
	if err != nil {
		log.Err(err)
		return err
	}

	if err = ExecContextRetry(ctx, db, counterTable, []string{}); err != nil {
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
		rows   *sql.Rows
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err = QueryContextRetry(ctx, db, "SELECT mname, val from gauge")
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := QueryContextRetry(ctx, db, "SELECT mname, val from counter")
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
	var (
		result = getGaugeCounter{name: "gauge"}
		err    error
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = QueryRowContextRetry(ctx, d.DBconn, "SELECT val from gauge where mname = $1", []string{name}, &result)
	if err != nil {
		return -1, false
	}

	return Gauge(result.gauge), true
}

func (d *DBstore) GetGauges() map[string]Gauge {
	var (
		result = make(map[string]Gauge)
		err    error
		rows   *sql.Rows
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err = QueryContextRetry(ctx, d.DBconn, "SELECT mname, val from gauge")
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
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	insertUpdate := `INSERT INTO gauge (mname, val) VALUES ($1, $2)
						ON CONFLICT (mname)
						DO UPDATE SET val = $3
						WHERE gauge.mname = $4`

	err = ExecContextRetry(ctx, d.DBconn, insertUpdate, []string{name, fmt.Sprintf("%v", value),
		fmt.Sprintf("%v", value), name})
	if err != nil {
		log.Err(err)
		panic(err)
	}
}

func (d *DBstore) GetCounter(name string) (Counter, bool) {
	var (
		result = getGaugeCounter{name: "counter"}
		err    error
	)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = QueryRowContextRetry(ctx, d.DBconn, "SELECT val from counter where mname = $1", []string{name}, &result)
	if err != nil {
		return -1, false
	}

	return Counter(result.counter), true
}

func (d *DBstore) GetCounters() map[string]Counter {
	var (
		result = make(map[string]Counter)
		err    error
		rows   *sql.Rows
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err = QueryContextRetry(ctx, d.DBconn, "SELECT mname, val from counter")
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
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	curVal, ok := d.GetCounter(name)
	if !ok {
		curVal = 0
	}
	curVal = curVal + value

	insertUpdate := `INSERT INTO counter (mname, val) VALUES ($1, $2)
						ON CONFLICT (mname)
						DO UPDATE SET val = $3
						WHERE counter.mname = $4`

	err = ExecContextRetry(ctx, d.DBconn, insertUpdate, []string{name, fmt.Sprintf("%v", curVal),
		fmt.Sprintf("%v", curVal), name})
	if err != nil {
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

func (d *DBstore) UpdateMetricBatch(reqJSON []models.Metrics) error {
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := d.DBconn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	insertUpdateGauge := `INSERT INTO gauge (mname, val) VALUES ($1, $2)
						ON CONFLICT (mname)
						DO UPDATE SET val = $3
						WHERE gauge.mname = $4`

	for _, v := range reqJSON {
		if v.MType == "gauge" {
			if _, err := tx.ExecContext(ctx, insertUpdateGauge, v.ID, v.Value, v.Value, v.ID); err != nil {
				log.Err(err)
				return err
			}
		} else if v.MType == "counter" {
			d.UpdateCounter(v.ID, Counter(*v.Delta))
		} else {
			err := fmt.Errorf("can not get val for %v from reqJSON", v.ID)
			return err
		}
	}

	err = txCommitRetry(tx)
	if err != nil {
		log.Err(err)
		return err
	}

	return nil
}
