package main

const (
	productChartView = `drop view if exists po_product_view;
	CREATE VIEW po_product_view AS
    SELECT po.user_name,
           po.user_id AS user_id,
           sum(pd.quantity) AS quantity,
           p.color,
           p.name AS product_name,
           p.id AS product_id
      FROM po_products pd
           INNER JOIN
           pos po ON pd.po_id = po.id
           INNER JOIN
           products p ON pd.product_id = p.id
     WHERE po.status = "approved"
     GROUP BY p.id,
              po.user_id;`
)

func (db Db) initViews() {
	db.orm.Exec(productChartView)
}
