package types

type Detail struct {
	Code   int    `json:"code"`
	Status bool   `json:"status"`
	Msg    string `json:"msg"`
	Data   struct {
		Msg  string `json:"msg"`
		Code int    `json:"code"`
		Data struct {
			ShareInfo struct {
				ShareLinkUrl     string `json:"shareLinkUrl"`
				MiniShareLinkUrl string `json:"miniShareLinkUrl"`
				ThreeDLinkUrl    string `json:"threeDLinkUrl"`
			} `json:"shareInfo"`
			MultiSizeShowSpread []interface{} `json:"multiSizeShowSpread"`
			HitArkSetting       bool          `json:"hitArkSetting"`
			MyOwnDto            struct {
				HasAdded bool `json:"hasAdded"`
			} `json:"myOwnDto"`
			RelationRecommend []interface{} `json:"relationRecommend"`
			LastSold          struct {
				ShowRecord        bool   `json:"showRecord"`
				LastSoldCountText string `json:"lastSoldCountText"`
				JumpRecordFlag    bool   `json:"jumpRecordFlag"`
				SoldCountText     string `json:"soldCountText"`
				List              []struct {
					OrderSubTypeName string `json:"orderSubTypeName"`
					Price            int    `json:"price"`
					FormatTime       string `json:"formatTime"`
					Avatar           string `json:"avatar"`
					UserName         string `json:"userName"`
					PropertiesValues string `json:"propertiesValues"`
				} `json:"list"`
			} `json:"lastSold"`
			ImageAndText []interface{} `json:"imageAndText"`
			AppraiseInfo struct {
				Score           string `json:"score"`
				EntryName       string `json:"entryName"`
				IsScoreShow     bool   `json:"isScoreShow"`
				ContentsNum     int    `json:"contentsNum"`
				IsOnline        bool   `json:"isOnline"`
				FusionEnabled   bool   `json:"fusionEnabled"`
				DimensionScores []struct {
					Score string `json:"score"`
					Id    int    `json:"id"`
					Title string `json:"title"`
				} `json:"dimensionScores"`
				ScoreTitle  string `json:"scoreTitle"`
				ContentInfo struct {
					UserInfo struct {
						Icon     string `json:"icon"`
						UserName string `json:"userName"`
						UserId   int    `json:"userId"`
					} `json:"userInfo"`
					Content struct {
						ContentId int `json:"contentId"`
						Media     struct {
							Total int `json:"total"`
							List  []struct {
								Width     int    `json:"width"`
								MediaType string `json:"mediaType"`
								MediaId   int    `json:"mediaId"`
								Url       string `json:"url"`
								Height    int    `json:"height"`
							} `json:"list"`
						} `json:"media"`
						DpInfo struct {
							Score         int `json:"score"`
							SpuProperties struct {
								Values []struct {
									Name string `json:"name"`
									Id   int    `json:"id"`
								} `json:"values"`
								Key string `json:"key"`
							} `json:"spuProperties"`
						} `json:"dpInfo"`
						Title       string `json:"title"`
						ContentType int    `json:"contentType"`
						Content     string `json:"content"`
					} `json:"content"`
				} `json:"contentInfo"`
				EntryId int  `json:"entryId"`
				IsShow  bool `json:"isShow"`
			} `json:"appraiseInfo"`
			SpuGroupList struct {
				Type int `json:"type"`
				List []struct {
					Spu3D360ShowType int    `json:"spu3d360ShowType"`
					SpuId            int    `json:"spuId"`
					Title            string `json:"title"`
					LogoUrl          string `json:"logoUrl"`
					GoodsType        int    `json:"goodsType"`
				} `json:"list"`
			} `json:"spuGroupList"`
			ToolFeatures []struct {
				Key        string `json:"key"`
				Url        string `json:"url"`
				GuideText  string `json:"guideText,omitempty"`
				QuickTools struct {
					Title string `json:"title"`
					Tools []struct {
						Name string `json:"name"`
						Key  string `json:"key"`
						Url  string `json:"url"`
					} `json:"tools"`
				} `json:"quickTools,omitempty"`
			} `json:"toolFeatures"`
			FavoriteList []struct {
				IsAdded int `json:"isAdded"`
				SkuId   int `json:"skuId"`
			} `json:"favoriteList"`
			Rank struct {
				RankIds    []int  `json:"rankIds"`
				RankId     int    `json:"rankId"`
				RankType   string `json:"rankType"`
				ThemeType  int    `json:"themeType"`
				Title      string `json:"title"`
				Seq        int    `json:"seq"`
				CategoryId int    `json:"categoryId"`
			} `json:"rank"`
			SpuInfoSpread []interface{} `json:"spuInfoSpread"`
			ConfigInfo    struct {
				HasEducationSpu        bool `json:"hasEducationSpu"`
				RelationTrendNum       int  `json:"relationTrendNum"`
				PostageFreeRatioReach  bool `json:"postageFreeRatioReach"`
				LastSoldModuleChange   bool `json:"lastSoldModuleChange"`
				DefaultPropertyValueId int  `json:"defaultPropertyValueId"`
				SuperNewProduct        struct {
					NewFlag bool `json:"newFlag"`
					SpuId   int  `json:"spuId"`
					Track   struct {
						BlockContentTitle string `json:"block_content_title"`
					} `json:"track"`
				} `json:"superNewProduct"`
				SpecialLogicTypeMap struct {
				} `json:"specialLogicTypeMap"`
				HasNewUser               bool   `json:"hasNewUser"`
				HasArMakeup              bool   `json:"hasArMakeup"`
				WidePriceText            string `json:"widePriceText"`
				TrendDoubleColumnFlow    bool   `json:"trendDoubleColumnFlow"`
				SellerFlag               bool   `json:"sellerFlag"`
				DefaultPropertyValueFlag bool   `json:"defaultPropertyValueFlag"`
				ServerTime               int64  `json:"serverTime"`
				IsOnlyShowHitCspu        bool   `json:"isOnlyShowHitCspu"`
				FloatLayerGood           bool   `json:"floatLayerGood"`
				OldSizeInfo              string `json:"oldSizeInfo"`
				HasSizeTable             bool   `json:"hasSizeTable"`
				NewUserFlag              bool   `json:"newUserFlag"`
				UseLevelImageUrl         bool   `json:"useLevelImageUrl"`
				MultipleSpuShow          bool   `json:"multipleSpuShow"`
				SpuHasBids               bool   `json:"spuHasBids"`
				CspuFlag                 bool   `json:"cspuFlag"`
				FoldShowRowNum           int    `json:"foldShowRowNum"`
				DownFoldShowRowNum       int    `json:"downFoldShowRowNum"`
				SizeFlag                 int    `json:"sizeFlag"`
				FloatLayerType           int    `json:"floatLayerType"`
				SizeInfo                 string `json:"sizeInfo"`
				HasSizeAggregation       bool   `json:"hasSizeAggregation"`
				ComparedRecommendRow     int    `json:"comparedRecommendRow"`
			} `json:"configInfo"`
			FloorsModel []struct {
				Models []string `json:"models"`
				Title  string   `json:"title"`
			} `json:"floorsModel"`
			ProductAdditionModel struct {
				AdditionSpu struct {
					ServiceBlocks []struct {
						AdditionType int           `json:"additionType"`
						ServiceItems []interface{} `json:"serviceItems"`
						PreviewType  int           `json:"previewType,omitempty"`
					} `json:"serviceBlocks"`
				} `json:"additionSpu"`
				HasAdditionFlag bool `json:"hasAdditionFlag"`
			} `json:"productAdditionModel"`
			Brand struct {
				BrandName                     string   `json:"brandName"`
				BrandPostCount                int      `json:"brandPostCount"`
				BrandRouterTab                int      `json:"brandRouterTab"`
				Icon                          string   `json:"icon"`
				PutOnTimeInSevenCount         int      `json:"putOnTimeInSevenCount"`
				BrandTagTextList              []string `json:"brandTagTextList"`
				CoverUrl                      string   `json:"coverUrl"`
				PutOnTimeInSeven              string   `json:"putOnTimeInSeven"`
				BrandOfSpuCount               int      `json:"brandOfSpuCount"`
				BrandOfSpuCountText           string   `json:"brandOfSpuCountText"`
				BrandId                       int      `json:"brandId"`
				TipList                       []string `json:"tipList"`
				BrandPostCountText            string   `json:"brandPostCountText"`
				StyleList                     []string `json:"styleList"`
				BrandSpuUserFavoriteCount     int      `json:"brandSpuUserFavoriteCount"`
				BrandLogo                     string   `json:"brandLogo"`
				BrandSpuUserFavoriteCountText string   `json:"brandSpuUserFavoriteCountText"`
				EnterText                     string   `json:"enterText"`
			} `json:"brand"`
			Image struct {
				ImageLayoutType int `json:"imageLayoutType"`
				SpuImage        struct {
					ArSkuIdRelation []struct {
						PropertyValueId int  `json:"propertyValueId"`
						SpuId           int  `json:"spuId"`
						SupportFlag     bool `json:"supportFlag"`
						SkuId           int  `json:"skuId"`
					} `json:"arSkuIdRelation"`
					Images []struct {
						PropertyValueId int    `json:"propertyValueId"`
						Url             string `json:"url"`
						ImgType         int    `json:"imgType"`
						Desc            string `json:"desc"`
					} `json:"images"`
					ColorBlockImages  []interface{} `json:"colorBlockImages"`
					MoreMaterialFlag  int           `json:"moreMaterialFlag"`
					SupportColorBlock int           `json:"supportColorBlock"`
					SupportPanorama   int           `json:"supportPanorama"`
					ThreeDimension    struct {
					} `json:"threeDimension"`
				} `json:"spuImage"`
			} `json:"image"`
			NewUserGuide []struct {
				Code  string `json:"code"`
				Value string `json:"value"`
			} `json:"newUserGuide"`
			Item struct {
				FloorPrice      int    `json:"floorPrice"`
				OriginalPrice   int    `json:"originalPrice"`
				SaleInventoryNo string `json:"saleInventoryNo"`
				Type            int    `json:"type"`
				TaxInfo         struct {
					IncludeTax bool `json:"includeTax"`
				} `json:"taxInfo"`
				IncludeTax    bool `json:"includeTax"`
				ActivityPrice int  `json:"activityPrice"`
				Price         int  `json:"price"`
				UpperPrice    int  `json:"upperPrice"`
				MaxPrice      int  `json:"maxPrice"`
				IntentPrice   int  `json:"intentPrice"`
				HaveBrandBid  bool `json:"haveBrandBid"`
				TradeType     int  `json:"tradeType"`
				SkuId         int  `json:"skuId"`
			} `json:"item"`
			QuestionAndAnswer struct {
				CanAsk        int    `json:"canAsk"`
				Total         int    `json:"total"`
				ShowMyQa      int    `json:"showMyQa"`
				TotalText     string `json:"totalText"`
				AskedRecently bool   `json:"askedRecently"`
				AggregationQa int    `json:"aggregationQa"`
				ShowType      int    `json:"showType"`
				QuestionList  []struct {
					AiAnswerNumber int `json:"aiAnswerNumber"`
					TopType        int `json:"topType"`
					QaAnswerList   []struct {
						AnswerOaId   int    `json:"answerOaId"`
						UserAvatar   string `json:"userAvatar"`
						CanAppend    int    `json:"canAppend"`
						UserName     string `json:"userName"`
						UsefulNumber int    `json:"usefulNumber"`
						UserId       int    `json:"userId"`
						Content      string `json:"content"`
						UserBuyTime  string `json:"userBuyTime,omitempty"`
						CreateTime   string `json:"createTime"`
						SourceType   int    `json:"sourceType"`
						QaQuestionId int    `json:"qaQuestionId"`
						Anonymous    int    `json:"anonymous"`
						SpuId        int    `json:"spuId"`
						Id           int    `json:"id"`
					} `json:"qaAnswerList"`
					UserAvatar       string `json:"userAvatar"`
					ShowStatus       int    `json:"showStatus"`
					CanAnswer        int    `json:"canAnswer"`
					UserName         string `json:"userName"`
					Type             int    `json:"type"`
					UserId           int    `json:"userId"`
					Content          string `json:"content"`
					CreateTime       string `json:"createTime"`
					HasFolded        bool   `json:"hasFolded"`
					ReadAnswerNumber int    `json:"readAnswerNumber"`
					Anonymous        int    `json:"anonymous"`
					SpuId            int    `json:"spuId"`
					ReviewStatus     int    `json:"reviewStatus"`
					Id               int    `json:"id"`
					AnswerNumber     int    `json:"answerNumber"`
					AiTagId          int    `json:"aiTagId"`
					Status           int    `json:"status"`
				} `json:"questionList"`
			} `json:"questionAndAnswer"`
			SpuIntroductionSpread []struct {
				Images []struct {
					Width     int           `json:"width"`
					SpuModels []interface{} `json:"spuModels"`
					Text      string        `json:"text"`
					Height    int           `json:"height"`
				} `json:"images"`
				GlobalImageList []interface{} `json:"globalImageList"`
				ImageGap        bool          `json:"imageGap"`
				Type            string        `json:"type"`
				ContentType     string        `json:"contentType"`
				ContentName     string        `json:"contentName"`
			} `json:"spuIntroductionSpread"`
			FaceShapeMeasureSpread []interface{} `json:"faceShapeMeasureSpread"`
			FreezingPointInfo      struct {
				HasDiscountActivity         bool   `json:"hasDiscountActivity"`
				CountDownBackUrl            string `json:"countDownBackUrl"`
				InFreezingPoint             bool   `json:"inFreezingPoint"`
				Price                       int    `json:"price"`
				FreezingPointStartTimeStamp int    `json:"freezingPointStartTimeStamp"`
				FreezingPointEndTimeStamp   int    `json:"freezingPointEndTimeStamp"`
				CurrentTimeStamp            int64  `json:"currentTimeStamp"`
				IconUrl                     string `json:"iconUrl"`
				FreezingPointPrice          int    `json:"freezingPointPrice"`
			} `json:"freezingPointInfo"`
			MainSkus         []interface{} `json:"mainSkus"`
			IdentifyBranding struct {
				Images []struct {
					Width     int           `json:"width"`
					SpuModels []interface{} `json:"spuModels"`
					Url       string        `json:"url"`
					JumpUrl   string        `json:"jumpUrl"`
					Height    int           `json:"height"`
				} `json:"images"`
				GlobalImageList []interface{} `json:"globalImageList"`
				ImageGap        bool          `json:"imageGap"`
			} `json:"identifyBranding"`
			AbnormalSceneType      int           `json:"abnormalSceneType"`
			PhysicalContrastSpread []interface{} `json:"physicalContrastSpread"`
			NewBrand               struct {
				CoverUrl     string        `json:"coverUrl"`
				BrandingDesc string        `json:"brandingDesc"`
				ArrowStyle   int           `json:"arrowStyle"`
				Width        int           `json:"width"`
				Id           int           `json:"id"`
				List         []interface{} `json:"list"`
				JumpUrl      string        `json:"jumpUrl"`
				Height       int           `json:"height"`
			} `json:"newBrand"`
			SizeInfo struct {
				SizeTemplate struct {
					BrandLogoUrl   string `json:"brandLogoUrl"`
					HitRowNum      int    `json:"hitRowNum"`
					HitColNum      int    `json:"hitColNum"`
					OffsetCodeType int    `json:"offsetCodeType"`
					Name           string `json:"name"`
					HasSizeTable   bool   `json:"hasSizeTable"`
					Title          string `json:"title"`
					List           []struct {
						SizeKey   string `json:"sizeKey"`
						SizeValue string `json:"sizeValue"`
					} `json:"list"`
					DefaultRowNum int    `json:"defaultRowNum"`
					Tips          string `json:"tips"`
					HasSelected   bool   `json:"hasSelected"`
					HasFold       bool   `json:"hasFold"`
				} `json:"sizeTemplate"`
			} `json:"sizeInfo"`
			Detail struct {
				FitId            int    `json:"fitId"`
				BizType          int    `json:"bizType"`
				DefaultSkuId     int    `json:"defaultSkuId"`
				Title            string `json:"title"`
				CategoryName     string `json:"categoryName"`
				LogoUrl          string `json:"logoUrl"`
				IsShow           int    `json:"isShow"`
				GoodsType        int    `json:"goodsType"`
				SubTitle         string `json:"subTitle"`
				ArticleNumber    string `json:"articleNumber"`
				BrandId          int    `json:"brandId"`
				SpuId            int    `json:"spuId"`
				Level2CategoryId int    `json:"level2CategoryId"`
				SellDate         string `json:"sellDate"`
				SourceName       string `json:"sourceName"`
				IsSelf           int    `json:"isSelf"`
				CategoryId       int    `json:"categoryId"`
				Level1CategoryId int    `json:"level1CategoryId"`
				AuthPrice        int    `json:"authPrice"`
				Desc             string `json:"desc"`
				Status           int    `json:"status"`
			} `json:"detail"`
			MainConfigInfo struct {
				HasEducationSpu       bool `json:"hasEducationSpu"`
				IsOnlyShowHitCspu     bool `json:"isOnlyShowHitCspu"`
				FloatLayerGood        bool `json:"floatLayerGood"`
				PostageFreeRatioReach bool `json:"postageFreeRatioReach"`
				LastSoldModuleChange  bool `json:"lastSoldModuleChange"`
				HasSizeTable          bool `json:"hasSizeTable"`
				NewUserFlag           bool `json:"newUserFlag"`
				UseLevelImageUrl      bool `json:"useLevelImageUrl"`
				MultipleSpuShow       bool `json:"multipleSpuShow"`
				SpuHasBids            bool `json:"spuHasBids"`
				SpecialLogicTypeMap   struct {
				} `json:"specialLogicTypeMap"`
				HasNewUser               bool  `json:"hasNewUser"`
				HasArMakeup              bool  `json:"hasArMakeup"`
				CspuFlag                 bool  `json:"cspuFlag"`
				TrendDoubleColumnFlow    bool  `json:"trendDoubleColumnFlow"`
				SellerFlag               bool  `json:"sellerFlag"`
				DefaultPropertyValueFlag bool  `json:"defaultPropertyValueFlag"`
				FloatLayerType           int   `json:"floatLayerType"`
				ServerTime               int64 `json:"serverTime"`
			} `json:"mainConfigInfo"`
			BrandStorySpread []struct {
				Images []struct {
					Width     int           `json:"width"`
					SpuModels []interface{} `json:"spuModels"`
					Text      string        `json:"text"`
					Height    int           `json:"height"`
				} `json:"images"`
				GlobalImageList []interface{} `json:"globalImageList"`
				ImageGap        bool          `json:"imageGap"`
				Type            string        `json:"type"`
				ContentType     string        `json:"contentType"`
				ContentName     string        `json:"contentName"`
			} `json:"brandStorySpread"`
			CommonQuestionRes struct {
				DetailType            int `json:"detailType"`
				CategoryLevel         int `json:"categoryLevel"`
				Level2CategoryId      int `json:"level2CategoryId"`
				FinalCategoryId       int `json:"finalCategoryId"`
				QuestionAndAnswerList []struct {
					Question string `json:"question"`
					Answer   string `json:"answer"`
				} `json:"questionAndAnswerList"`
				RuleId           string `json:"ruleId"`
				JumpUrl          string `json:"jumpUrl"`
				Level1CategoryId int    `json:"level1CategoryId"`
				CategoryId       int    `json:"categoryId"`
			} `json:"commonQuestionRes"`
			Evaluate struct {
				Sizes      []interface{} `json:"sizes"`
				LimitCount int           `json:"limitCount"`
				Count      int           `json:"count"`
				Type       int           `json:"type"`
			} `json:"evaluate"`
			Skus []struct {
				SpuId      int `json:"spuId"`
				SkuId      int `json:"skuId"`
				AuthPrice  int `json:"authPrice"`
				Properties []struct {
					Level           int `json:"level"`
					PropertyValueId int `json:"propertyValueId"`
				} `json:"properties"`
				LogoUrl string `json:"logoUrl"`
				Status  int    `json:"status"`
			} `json:"skus"`
			NewService struct {
				ArrowStyle int `json:"arrowStyle"`
				List       []struct {
					CheckTarget       string `json:"checkTarget"`
					SnapShotTimestamp int64  `json:"snapShotTimestamp"`
					Title             string `json:"title"`
					Type              int    `json:"type"`
				} `json:"list"`
			} `json:"newService"`
			ModelSequence []struct {
				Title string `json:"title,omitempty"`
				Key   string `json:"key"`
				Value []struct {
					Key   string `json:"key"`
					Title string `json:"title,omitempty"`
				} `json:"value,omitempty"`
			} `json:"modelSequence"`
			SpuWearSpread  []interface{} `json:"spuWearSpread"`
			SaleProperties struct {
				List []struct {
					ShowValue       int    `json:"showValue"`
					Level           int    `json:"level"`
					CustomValue     string `json:"customValue"`
					PropertyValueId int    `json:"propertyValueId"`
					Name            string `json:"name"`
					Sort            int    `json:"sort"`
					PropertyId      int    `json:"propertyId"`
					Value           string `json:"value"`
					FontProperty    bool   `json:"fontProperty"`
					DefinitionId    int    `json:"definitionId"`
				} `json:"list"`
			} `json:"saleProperties"`
			SpuShowSpread []struct {
				Images []struct {
					Width     int           `json:"width"`
					SpuModels []interface{} `json:"spuModels"`
					Url       string        `json:"url"`
					Height    int           `json:"height"`
				} `json:"images"`
				GlobalImageList []interface{} `json:"globalImageList"`
				ImageGap        bool          `json:"imageGap"`
				Type            string        `json:"type"`
				ContentType     string        `json:"contentType"`
				ContentName     string        `json:"contentName"`
			} `json:"spuShowSpread"`
			FrameReferenceSpread []interface{} `json:"frameReferenceSpread"`
			Spu3D360List         struct {
				List []struct {
					ShowType int    `json:"showType"`
					SpuId    int    `json:"spuId"`
					LogoUrl  string `json:"logoUrl"`
				} `json:"list"`
			} `json:"spu3d360List"`
			AbnormalSceneToast    string        `json:"abnormalSceneToast"`
			CapacityDiagramSpread []interface{} `json:"capacityDiagramSpread"`
			Button                struct {
				BuyNow struct {
					Title  string `json:"title"`
					IsShow int    `json:"isShow"`
				} `json:"buyNow"`
				Sell struct {
					Title  string `json:"title"`
					IsShow int    `json:"isShow"`
				} `json:"sell"`
				WantBuy struct {
					Title  string `json:"title"`
					IsShow int    `json:"isShow"`
				} `json:"wantBuy"`
				Share struct {
					IsShow int `json:"isShow"`
				} `json:"share"`
				Kefu struct {
					IsShow int `json:"isShow"`
				} `json:"kefu"`
				Favorite struct {
					Title  string `json:"title"`
					IsShow int    `json:"isShow"`
				} `json:"favorite"`
			} `json:"button"`
			NewEvaluate struct {
				LimitCount   int `json:"limitCount"`
				Satisfaction struct {
					CommonSize string `json:"commonSize"`
					Count      int    `json:"count"`
					Type       int    `json:"type"`
					RateMsg    string `json:"rateMsg"`
				} `json:"satisfaction"`
				Items []struct {
					Amount   int           `json:"amount"`
					SubItems []interface{} `json:"subItems"`
					Name     string        `json:"name"`
				} `json:"items"`
				SignModel struct {
					SubRatio []interface{} `json:"subRatio"`
					Summery  string        `json:"summery"`
					Ratio    []struct {
						Amount     int           `json:"amount"`
						SubItems   []interface{} `json:"subItems"`
						Percentage string        `json:"percentage"`
						Name       string        `json:"name"`
					} `json:"ratio"`
				} `json:"signModel"`
			} `json:"newEvaluate"`
			BaseProperties struct {
				BrandList []struct {
					BrandName string `json:"brandName"`
					BrandId   int    `json:"brandId"`
				} `json:"brandList"`
				List []struct {
					Value           string `json:"value"`
					Key             string `json:"key"`
					PropertyValueId int    `json:"propertyValueId,omitempty"`
				} `json:"list"`
				Title   string `json:"title"`
				PkModel struct {
					PkFlag bool `json:"pkFlag"`
				} `json:"pkModel"`
			} `json:"baseProperties"`
			SpecificationCourseSpread []interface{} `json:"specificationCourseSpread"`
			ProductSoldOutModel       struct {
			} `json:"productSoldOutModel"`
			KeyProperties    []interface{} `json:"keyProperties"`
			PlatformBranding struct {
				Images []struct {
					Width     int           `json:"width"`
					SpuModels []interface{} `json:"spuModels"`
					Url       string        `json:"url"`
					JumpUrl   string        `json:"jumpUrl"`
					Height    int           `json:"height"`
				} `json:"images"`
				GlobalImageList []interface{} `json:"globalImageList"`
				ImageGap        bool          `json:"imageGap"`
			} `json:"platformBranding"`
			SpuDetailSpread []struct {
				Images []struct {
					Width     int           `json:"width"`
					SpuModels []interface{} `json:"spuModels"`
					Url       string        `json:"url"`
					Height    int           `json:"height"`
				} `json:"images"`
				GlobalImageList []interface{} `json:"globalImageList"`
				ImageGap        bool          `json:"imageGap"`
				Type            string        `json:"type"`
				ContentType     string        `json:"contentType"`
				ContentName     string        `json:"contentName"`
			} `json:"spuDetailSpread"`
			HasWearRecommend    bool `json:"hasWearRecommend"`
			SpuLabelSummaryList []struct {
				LabelInfo []struct {
					LabelContent string `json:"labelContent"`
					TipsContent  string `json:"tipsContent"`
					LabelType    int    `json:"labelType"`
					TipsIcon     string `json:"tipsIcon"`
				} `json:"labelInfo"`
				Position int `json:"position"`
			} `json:"spuLabelSummaryList"`
			HasBrandOrArtist     string        `json:"hasBrandOrArtist"`
			RichTextSpreadSwitch bool          `json:"richTextSpreadSwitch"`
			AdvMidAbFlag         bool          `json:"advMidAbFlag"`
			DesignConceptSpread  []interface{} `json:"designConceptSpread"`
			SpuVideoSpread       []interface{} `json:"spuVideoSpread"`
			SizeWearEffectSpread []interface{} `json:"sizeWearEffectSpread"`
			NewDiscountInfo      struct {
				DiscountTags    []interface{} `json:"discountTags"`
				OptimalDiscount struct {
					OptimalPriceWithReachable int `json:"optimalPriceWithReachable"`
				} `json:"optimalDiscount"`
				CanOpenDiscountFloat bool `json:"canOpenDiscountFloat"`
			} `json:"newDiscountInfo"`
			BuyerReading struct {
				ShowNum        int    `json:"showNum"`
				ComponentIndex int    `json:"componentIndex"`
				Title          string `json:"title"`
				ContentList    []struct {
					Title   string `json:"title"`
					Content string `json:"content"`
				} `json:"contentList"`
			} `json:"buyerReading"`
			Attention struct {
				Images []struct {
					Width     int           `json:"width"`
					SpuModels []interface{} `json:"spuModels"`
					Text      string        `json:"text"`
					Height    int           `json:"height"`
				} `json:"images"`
				GlobalImageList []interface{} `json:"globalImageList"`
				ImageGap        bool          `json:"imageGap"`
				Type            string        `json:"type"`
				ContentType     string        `json:"contentType"`
				ContentName     string        `json:"contentName"`
			} `json:"attention"`
			ActivityForm struct {
				ImageUrl           string `json:"imageUrl"`
				Subhead            string `json:"subhead"`
				SpareText          string `json:"spareText"`
				CurrentTimeStamp   int64  `json:"currentTimeStamp"`
				Type               int    `json:"type"`
				Title              string `json:"title"`
				CountDownText      string `json:"countDownText"`
				RouterUrl          string `json:"routerUrl"`
				BackgroundImageUrl string `json:"backgroundImageUrl"`
			} `json:"activityForm"`
			KfInfo struct {
				KfRobotoId int    `json:"kfRobotoId"`
				KfGroupId  int    `json:"kfGroupId"`
				IsShowKf   int    `json:"isShowKf"`
				SourceName string `json:"sourceName"`
				KfRobotId  int    `json:"kfRobotId"`
			} `json:"kfInfo"`
			FavoriteCount struct {
				Count   int    `json:"count"`
				Content string `json:"content"`
			} `json:"favoriteCount"`
		} `json:"data"`
		Error  bool `json:"error"`
		Status int  `json:"status"`
	} `json:"data"`
	Ts int64 `json:"ts"`
}
