package lib

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"

	"github.com/teaql/teaql-golang/core"
	provider "github.com/teaql/teaql-golang/provider/sqlite"
	"github.com/teaql/teaql-golang/runtime"
	teaql_sql "github.com/teaql/teaql-golang/sql"

	"order-management-service-core-workspace/lib/commerce_platform"
	"order-management-service-core-workspace/lib/customer"
	"order-management-service-core-workspace/lib/customer_order"
	"order-management-service-core-workspace/lib/order_line"
	"order-management-service-core-workspace/lib/order_search_preset"
	"order-management-service-core-workspace/lib/order_status"
	"order-management-service-core-workspace/lib/product"
)

func Module() *runtime.RuntimeModule {
	module := runtime.NewRuntimeModule()
	module.Entity(
		core.NewEntityDescriptor("Commerce Platform").
			TableName("commerce_platform_data").Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id()).Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name")).Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time")).Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time")).Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version()).Relation(core.NewRelationDescriptor("customerList", "Customer").LocalKey("id").ForeignKey("commerce_platform_id").Many()).Relation(core.NewRelationDescriptor("orderStatusList", "Order Status").LocalKey("id").ForeignKey("commerce_platform_id").Many()).Relation(core.NewRelationDescriptor("customerOrderList", "Customer Order").LocalKey("id").ForeignKey("commerce_platform_id").Many()).Relation(core.NewRelationDescriptor("productList", "Product").LocalKey("id").ForeignKey("commerce_platform_id").Many()).Relation(core.NewRelationDescriptor("orderLineList", "Order Line").LocalKey("id").ForeignKey("commerce_platform_id").Many()).Relation(core.NewRelationDescriptor("orderSearchPresetList", "Order Search Preset").LocalKey("id").ForeignKey("commerce_platform_id").Many()),
	)
	module.Entity(
		core.NewEntityDescriptor("Customer").
			TableName("customer_data").Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id()).Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name")).Property(core.NewPropertyDescriptor("email", core.TypeText).ColumnName("email")).Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time")).Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time")).Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version()).Property(core.NewPropertyDescriptor("commerce_platform_id", core.TypeU64).ColumnName("commerce_platform")).Relation(core.NewRelationDescriptor("customerOrderList", "Customer Order").LocalKey("id").ForeignKey("customer_id").Many()),
	)
	module.Entity(
		core.NewEntityDescriptor("Order Status").
			TableName("order_status_data").Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id()).Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name")).Property(core.NewPropertyDescriptor("code", core.TypeText).ColumnName("code")).Property(core.NewPropertyDescriptor("color", core.TypeText).ColumnName("color")).Property(core.NewPropertyDescriptor("display_order", core.TypeDecimal).ColumnName("display_order")).Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version()).Property(core.NewPropertyDescriptor("commerce_platform_id", core.TypeU64).ColumnName("commerce_platform")).Relation(core.NewRelationDescriptor("customerOrderList", "Customer Order").LocalKey("id").ForeignKey("status_id").Many()),
	)
	module.Entity(
		core.NewEntityDescriptor("Customer Order").
			TableName("customer_order_data").Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id()).Property(core.NewPropertyDescriptor("order_number", core.TypeText).ColumnName("order_number")).Property(core.NewPropertyDescriptor("order_date", core.TypeDate).ColumnName("order_date")).Property(core.NewPropertyDescriptor("total_amount", core.TypeDecimal).ColumnName("total_amount")).Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time")).Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time")).Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version()).Property(core.NewPropertyDescriptor("status_id", core.TypeU64).ColumnName("status")).Property(core.NewPropertyDescriptor("customer_id", core.TypeU64).ColumnName("customer")).Property(core.NewPropertyDescriptor("commerce_platform_id", core.TypeU64).ColumnName("commerce_platform")).Relation(core.NewRelationDescriptor("orderLineList", "Order Line").LocalKey("id").ForeignKey("customer_order_id").Many()),
	)
	module.Entity(
		core.NewEntityDescriptor("Product").
			TableName("product_data").Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id()).Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name")).Property(core.NewPropertyDescriptor("sku", core.TypeText).ColumnName("sku")).Property(core.NewPropertyDescriptor("image_url", core.TypeText).ColumnName("image_url")).Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time")).Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time")).Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version()).Property(core.NewPropertyDescriptor("commerce_platform_id", core.TypeU64).ColumnName("commerce_platform")).Relation(core.NewRelationDescriptor("orderLineList", "Order Line").LocalKey("id").ForeignKey("product_id").Many()),
	)
	module.Entity(
		core.NewEntityDescriptor("Order Line").
			TableName("order_line_data").Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id()).Property(core.NewPropertyDescriptor("product_name", core.TypeText).ColumnName("product_name")).Property(core.NewPropertyDescriptor("sku", core.TypeText).ColumnName("sku")).Property(core.NewPropertyDescriptor("quantity", core.TypeI64).ColumnName("quantity")).Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time")).Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version()).Property(core.NewPropertyDescriptor("customer_order_id", core.TypeU64).ColumnName("customer_order")).Property(core.NewPropertyDescriptor("product_id", core.TypeU64).ColumnName("product")).Property(core.NewPropertyDescriptor("commerce_platform_id", core.TypeU64).ColumnName("commerce_platform")),
	)
	module.Entity(
		core.NewEntityDescriptor("Order Search Preset").
			TableName("order_search_preset_data").Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id()).Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name")).Property(core.NewPropertyDescriptor("filter_json", core.TypeText).ColumnName("filter_json")).Property(core.NewPropertyDescriptor("request_id", core.TypeText).ColumnName("request_id")).Property(core.NewPropertyDescriptor("owner_user_id", core.TypeText).ColumnName("owner_user_id")).Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time")).Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time")).Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version()).Property(core.NewPropertyDescriptor("commerce_platform_id", core.TypeU64).ColumnName("commerce_platform")),
	)
	return module
}

func ModuleWithBehaviors() *runtime.RuntimeModule {
	module := runtime.NewRuntimeModule()
	module.EntityWithBehavior(
		core.NewEntityDescriptor("Commerce Platform").
			TableName("commerce_platform_data").Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id()).Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name")).Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time")).Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time")).Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version()).Relation(core.NewRelationDescriptor("customerList", "Customer").LocalKey("id").ForeignKey("commerce_platform_id").Many()).Relation(core.NewRelationDescriptor("orderStatusList", "Order Status").LocalKey("id").ForeignKey("commerce_platform_id").Many()).Relation(core.NewRelationDescriptor("customerOrderList", "Customer Order").LocalKey("id").ForeignKey("commerce_platform_id").Many()).Relation(core.NewRelationDescriptor("productList", "Product").LocalKey("id").ForeignKey("commerce_platform_id").Many()).Relation(core.NewRelationDescriptor("orderLineList", "Order Line").LocalKey("id").ForeignKey("commerce_platform_id").Many()).Relation(core.NewRelationDescriptor("orderSearchPresetList", "Order Search Preset").LocalKey("id").ForeignKey("commerce_platform_id").Many()),
		&commerce_platform.CommercePlatformBehavior{},
	)
	module.EntityWithBehavior(
		core.NewEntityDescriptor("Customer").
			TableName("customer_data").Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id()).Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name")).Property(core.NewPropertyDescriptor("email", core.TypeText).ColumnName("email")).Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time")).Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time")).Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version()).Property(core.NewPropertyDescriptor("commerce_platform_id", core.TypeU64).ColumnName("commerce_platform")).Relation(core.NewRelationDescriptor("customerOrderList", "Customer Order").LocalKey("id").ForeignKey("customer_id").Many()),
		&customer.CustomerBehavior{},
	)
	module.EntityWithBehavior(
		core.NewEntityDescriptor("Order Status").
			TableName("order_status_data").Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id()).Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name")).Property(core.NewPropertyDescriptor("code", core.TypeText).ColumnName("code")).Property(core.NewPropertyDescriptor("color", core.TypeText).ColumnName("color")).Property(core.NewPropertyDescriptor("display_order", core.TypeDecimal).ColumnName("display_order")).Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version()).Property(core.NewPropertyDescriptor("commerce_platform_id", core.TypeU64).ColumnName("commerce_platform")).Relation(core.NewRelationDescriptor("customerOrderList", "Customer Order").LocalKey("id").ForeignKey("status_id").Many()),
		&order_status.OrderStatusBehavior{},
	)
	module.EntityWithBehavior(
		core.NewEntityDescriptor("Customer Order").
			TableName("customer_order_data").Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id()).Property(core.NewPropertyDescriptor("order_number", core.TypeText).ColumnName("order_number")).Property(core.NewPropertyDescriptor("order_date", core.TypeDate).ColumnName("order_date")).Property(core.NewPropertyDescriptor("total_amount", core.TypeDecimal).ColumnName("total_amount")).Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time")).Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time")).Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version()).Property(core.NewPropertyDescriptor("status_id", core.TypeU64).ColumnName("status")).Property(core.NewPropertyDescriptor("customer_id", core.TypeU64).ColumnName("customer")).Property(core.NewPropertyDescriptor("commerce_platform_id", core.TypeU64).ColumnName("commerce_platform")).Relation(core.NewRelationDescriptor("orderLineList", "Order Line").LocalKey("id").ForeignKey("customer_order_id").Many()),
		&customer_order.CustomerOrderBehavior{},
	)
	module.EntityWithBehavior(
		core.NewEntityDescriptor("Product").
			TableName("product_data").Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id()).Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name")).Property(core.NewPropertyDescriptor("sku", core.TypeText).ColumnName("sku")).Property(core.NewPropertyDescriptor("image_url", core.TypeText).ColumnName("image_url")).Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time")).Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time")).Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version()).Property(core.NewPropertyDescriptor("commerce_platform_id", core.TypeU64).ColumnName("commerce_platform")).Relation(core.NewRelationDescriptor("orderLineList", "Order Line").LocalKey("id").ForeignKey("product_id").Many()),
		&product.ProductBehavior{},
	)
	module.EntityWithBehavior(
		core.NewEntityDescriptor("Order Line").
			TableName("order_line_data").Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id()).Property(core.NewPropertyDescriptor("product_name", core.TypeText).ColumnName("product_name")).Property(core.NewPropertyDescriptor("sku", core.TypeText).ColumnName("sku")).Property(core.NewPropertyDescriptor("quantity", core.TypeI64).ColumnName("quantity")).Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time")).Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version()).Property(core.NewPropertyDescriptor("customer_order_id", core.TypeU64).ColumnName("customer_order")).Property(core.NewPropertyDescriptor("product_id", core.TypeU64).ColumnName("product")).Property(core.NewPropertyDescriptor("commerce_platform_id", core.TypeU64).ColumnName("commerce_platform")),
		&order_line.OrderLineBehavior{},
	)
	module.EntityWithBehavior(
		core.NewEntityDescriptor("Order Search Preset").
			TableName("order_search_preset_data").Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id()).Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name")).Property(core.NewPropertyDescriptor("filter_json", core.TypeText).ColumnName("filter_json")).Property(core.NewPropertyDescriptor("request_id", core.TypeText).ColumnName("request_id")).Property(core.NewPropertyDescriptor("owner_user_id", core.TypeText).ColumnName("owner_user_id")).Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time")).Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time")).Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version()).Property(core.NewPropertyDescriptor("commerce_platform_id", core.TypeU64).ColumnName("commerce_platform")),
		&order_search_preset.OrderSearchPresetBehavior{},
	)
	return module
}

type schemaProviderAdapter struct {
	metadata runtime.MetadataStore
}

func (a *schemaProviderAdapter) GetEntity(name string) *core.EntityDescriptor {
	return a.metadata.Entity(name)
}

func ServiceRuntimeFromEnv() (*runtime.UserContext, error) {
	dbUrl := os.Getenv("ORDER_MANAGEMENT_SERVICE_CORE_DATABASE_URL")
	if dbUrl == "" {
		return nil, fmt.Errorf("missing environment variable ORDER_MANAGEMENT_SERVICE_CORE_DATABASE_URL")
	}

	db, err := sql.Open("sqlite3", dbUrl)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	module := ModuleWithBehaviors()
	context := module.IntoContext()

	dialect := &provider.SqliteDialect{}
	transport := provider.NewSqliteMutationExecutor(db)
	executor := teaql_sql.NewSqlDataServiceExecutor(dialect, transport, &schemaProviderAdapter{module.Metadata})
	if err := ensureGeneratedSchema(db, dialect, module.Metadata); err != nil {
		db.Close()
		return nil, err
	}

	context.InsertResource("dataService", executor)
	context.InsertResource("db", db)
	context.InsertResource("idGenerator", &sqlIdGenerator{db: db})

	return context, nil
}

func ensureGeneratedSchema(db *sql.DB, dialect teaql_sql.SqlDialect, metadata runtime.MetadataStore) error {
	for _, statement := range dialect.SchemaSetupSqls() {
		if _, err := db.Exec(statement); err != nil {
			return fmt.Errorf("initialize SQL dialect: %w", err)
		}
	}
	defaultDialect := &teaql_sql.DefaultSqlDialect{Dialect: dialect}
	for _, entity := range metadata.AllEntities() {
		statement, err := defaultDialect.CompileCreateTable(entity)
		if err != nil {
			return fmt.Errorf("compile schema for %s: %w", entity.Name, err)
		}
		if _, err := db.Exec(statement); err != nil {
			return fmt.Errorf("create table for %s: %w", entity.Name, err)
		}
		indexes, err := defaultDialect.SchemaIndexesSqls(entity)
		if err != nil {
			return fmt.Errorf("compile indexes for %s: %w", entity.Name, err)
		}
		for _, indexStatement := range indexes {
			if _, err := db.Exec(indexStatement); err != nil {
				return fmt.Errorf("create index for %s: %w", entity.Name, err)
			}
		}
	}
	return nil
}

type sqlIdGenerator struct {
	db *sql.DB
}

func (g *sqlIdGenerator) GenerateId(entity string) (uint64, error) {
	if _, err := g.db.Exec(`CREATE TABLE IF NOT EXISTS teaql_id_space (
		entity VARCHAR(255) PRIMARY KEY,
		next_id BIGINT NOT NULL
	)`); err != nil {
		return 0, err
	}
	var id int64
	err := g.db.QueryRow(`
		INSERT INTO teaql_id_space(entity, next_id) VALUES (?, 1000)
		ON CONFLICT(entity) DO UPDATE SET next_id = teaql_id_space.next_id + 1
		RETURNING next_id
	`, entity).Scan(&id)
	return uint64(id), err
}
