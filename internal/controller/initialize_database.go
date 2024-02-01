/*
Copyright 2021 Pachyderm.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"

	// postgres database drivers
	_ "github.com/lib/pq"
	aimlv1beta1 "github.com/pachyderm/openshift-operator/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *PachydermReconciler) initializePostgres(ctx context.Context, pd *aimlv1beta1.Pachyderm) error {
	initScript, err := r.loadPostgresInitQueries(ctx, pd)
	if err != nil {
		return err
	}

	adminPassword, err := r.getPostgresAdminPassword(ctx, pd)
	if err != nil {
		return err
	}

	dataSource := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		fmt.Sprintf("%s.%s", pd.Spec.Pachd.Postgres.Host, pd.Namespace),
		pd.Spec.Pachd.Postgres.Port,
		"postgres",
		adminPassword,
		pd.Spec.Pachd.Postgres.Database,
	)
	db, err := sql.Open("postgres", dataSource)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return err
	}

	var dexTableExists bool
	existQuery := "SELECT exists(SELECT datname FROM pg_catalog.pg_database WHERE datname = 'dex');"
	row := db.QueryRow(existQuery)
	if err := row.Scan(&dexTableExists); err != nil {
		return err
	}
	if !dexTableExists {
		for _, query := range initScript {
			_, err := db.ExecContext(ctx, query)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *PachydermReconciler) loadPostgresInitQueries(ctx context.Context, pd *aimlv1beta1.Pachyderm) ([]string, error) {
	initScripts := &corev1.ConfigMap{}
	initScriptsKey := types.NamespacedName{
		Name:      "postgres-init-scripts",
		Namespace: pd.Namespace,
	}
	if err := r.Get(ctx, initScriptsKey, initScripts); err != nil {
		return nil, err
	}

	script, ok := initScripts.Data["dex.sh"]
	if !ok {
		return nil, errors.New("init script dex.sh not found")
	}

	script = strings.ReplaceAll(script, "$POSTGRES_USER", pd.Spec.Pachd.Postgres.User)
	re := regexp.MustCompile(`(?m)^\s*([A-Z].*)\;$`)
	return re.FindAllString(script, -1), nil
}

func (r *PachydermReconciler) getPostgresAdminPassword(ctx context.Context, pd *aimlv1beta1.Pachyderm) (string, error) {
	secret := &corev1.Secret{}
	secretKey := types.NamespacedName{
		Namespace: pd.Namespace,
		Name:      "postgres",
	}
	if err := r.Get(ctx, secretKey, secret); err != nil {
		return "", err
	}
	dbAdminPassword, ok := secret.Data["postgresql-postgres-password"]
	if ok {
		return string(dbAdminPassword), nil
	}

	dbAdminPassword, ok = secret.Data["postgres-password"]
	if ok {
		return string(dbAdminPassword), nil
	}

	dbAdminPassword, ok = secret.Data["postgresql-password"]
	if ok {
		return string(dbAdminPassword), nil
	}

	return "", errors.New("postgresql key \"postgresql-postgres-password\" missing")
}
