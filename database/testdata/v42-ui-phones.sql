-- V42 UI QA fixtures: deliberately varied phone profiles for visual regression.
INSERT INTO phone_numbers(country_code,calling_code,national_number,e164,number_type,verification_status,seo_status,spam_score,report_count,search_count,data_quality_score)
VALUES
('VN','84','901000001','+84901000001','mobile','verified','indexable',0.00,0,1250,98),
('VN','84','901000002','+84901000002','mobile','unverified','noindex',0.05,0,18,42),
('VN','84','901000003','+84901000003','mobile','verified','indexable',0.18,2,845,88),
('VN','84','901000004','+84901000004','mobile','disputed','noindex',0.72,8,2310,65),
('VN','84','901000005','+84901000005','mobile','unverified','noindex',0.95,21,7600,30),
('VN','84','2810000006','+842810000006','landline','verified','indexable',0.00,0,420,96),
('VN','84','2610000007','+842610000007','landline','verified','indexable',0.08,1,76,82),
('VN','84','901000008','+84901000008','mobile','pending','noindex',0.35,3,350,58),
('VN','84','901000009','+84901000009','mobile','disputed','noindex',0.61,6,910,40),
('VN','84','901000010','+84901000010','mobile','verified','indexable',0.12,1,15000,91)
ON CONFLICT(e164) DO UPDATE SET verification_status=EXCLUDED.verification_status,seo_status=EXCLUDED.seo_status,spam_score=EXCLUDED.spam_score,report_count=EXCLUDED.report_count,search_count=EXCLUDED.search_count,data_quality_score=EXCLUDED.data_quality_score;

INSERT INTO phone_identities(phone_number_id,kind,display_name,description,address_text,website_url,source_label,is_primary,is_public,confidence_score)
SELECT p.id,v.kind::identity_kind,v.display_name,v.description,v.address_text,v.website_url,'V42 UI fixture',TRUE,TRUE,v.confidence_score
FROM (VALUES
('+84901000001','business','Taxi Minh Anh','Doanh nghiệp vận tải đã xác minh.','Đà Lạt, Lâm Đồng','https://example.com/minh-anh',98.00),
('+84901000003','business','Điện máy An Phát','Cửa hàng điện máy địa phương.','Việt Nam','https://example.com/an-phat',88.00),
('+842810000006','organization','Tổng đài doanh nghiệp mẫu','Số bàn doanh nghiệp dùng để kiểm tra giao diện.','TP. Hồ Chí Minh',NULL,96.00),
('+842610000007','service','Dịch vụ địa phương mẫu','Hồ sơ có một cảnh báo cộng đồng.','Lâm Đồng',NULL,82.00),
('+84901000010','business','Doanh nghiệp tên rất dài để kiểm tra khả năng xuống dòng trên giao diện SoDienThoai.com','Kiểm tra hiển thị nội dung dài trên desktop và mobile.','Địa chỉ mẫu dài, Việt Nam','https://example.com/long-profile',91.00)
) AS v(e164,kind,display_name,description,address_text,website_url,confidence_score)
JOIN phone_numbers p ON p.e164=v.e164
WHERE NOT EXISTS(SELECT 1 FROM phone_identities i WHERE i.phone_number_id=p.id AND i.is_primary=TRUE);

INSERT INTO phone_sources(phone_number_id,source_type,source_ref,confidence)
SELECT p.id,'v42_ui_fixture','database/testdata/v42-ui-phones.sql',1.0 FROM phone_numbers p
WHERE (p.e164 LIKE '+849010000%' OR p.e164 IN('+842810000006','+842610000007'))
AND NOT EXISTS(SELECT 1 FROM phone_sources s WHERE s.phone_number_id=p.id AND s.source_type='v42_ui_fixture');
