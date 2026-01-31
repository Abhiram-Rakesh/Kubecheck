{-# LANGUAGE DeriveGeneric #-}
{-# LANGUAGE OverloadedStrings #-}

module Types
    ( K8sResource(..)
    , Metadata(..)
    , Spec(..)
    , PodTemplate(..)
    , PodSpec(..)
    , Container(..)
    , SecurityContext(..)
    , Resources(..)
    , ResourceRequirements(..)
    , Violation(..)
    , Severity(..)
    ) where

import Data.Aeson (FromJSON, ToJSON, Value)
import qualified Data.Aeson as A
import Data.Map.Strict (Map)
import Data.Text (Text)
import GHC.Generics (Generic)

-- | Severity levels for violations
data Severity
    = SeverityOK
    | SeverityWarn
    | SeverityError
    deriving (Eq, Show, Generic)

instance ToJSON Severity where
    toJSON SeverityOK    = A.String "OK"
    toJSON SeverityWarn  = A.String "WARN"
    toJSON SeverityError = A.String "ERROR"

-- | Validation violation
data Violation = Violation
    { severity :: Severity
    , message  :: Text
    , rule     :: Text
    } deriving (Eq, Show, Generic)

instance ToJSON Violation where
    toJSON (Violation sev msg r) =
        A.object [ "severity" A..= sev
                 , "message"  A..= msg
                 , "rule"     A..= r
                 ]

-- | Kubernetes resource
data K8sResource = K8sResource
    { apiVersion :: Text
    , kind       :: Text
    , metadata   :: Map Text Value
    , spec       :: Maybe Spec
    } deriving (Eq, Show, Generic)

instance FromJSON K8sResource where
    parseJSON = A.withObject "K8sResource" $ \v ->
        K8sResource
            <$> v A..: "apiVersion"
            <*> v A..: "kind"
            <*> v A..: "metadata"
            <*> v A..:? "spec"

-- | Metadata
data Metadata = Metadata
    { metadataName      :: Maybe Text
    , metadataNamespace :: Maybe Text
    } deriving (Eq, Show, Generic)

instance FromJSON Metadata

-- | Spec (simplified for common workload types)
data Spec = Spec
    { template   :: Maybe PodTemplate
    , containers :: Maybe [Container]
    } deriving (Eq, Show, Generic)

instance FromJSON Spec where
    parseJSON = A.withObject "Spec" $ \v ->
        Spec
            <$> v A..:? "template"
            <*> v A..:? "containers"

-- | Pod template
data PodTemplate = PodTemplate
    { podSpec :: PodSpec
    } deriving (Eq, Show, Generic)

instance FromJSON PodTemplate where
    parseJSON = A.withObject "PodTemplate" $ \v -> do
        ps <- v A..: "spec"
        return $ PodTemplate ps

-- | Pod spec
data PodSpec = PodSpec
    { podContainers :: [Container]
    } deriving (Eq, Show, Generic)

instance FromJSON PodSpec where
    parseJSON = A.withObject "PodSpec" $ \v ->
        PodSpec <$> v A..: "containers"

-- | Container specification
data Container = Container
    { containerName           :: Text
    , containerImage          :: Text
    , containerSecurityContext :: Maybe SecurityContext
    , containerResources      :: Maybe Resources
    } deriving (Eq, Show, Generic)

instance FromJSON Container where
    parseJSON = A.withObject "Container" $ \v ->
        Container
            <$> v A..: "name"
            <*> v A..: "image"
            <*> v A..:? "securityContext"
            <*> v A..:? "resources"

-- | Security context
data SecurityContext = SecurityContext
    { runAsNonRoot :: Maybe Bool
    , runAsUser    :: Maybe Int
    } deriving (Eq, Show, Generic)

instance FromJSON SecurityContext where
    parseJSON = A.withObject "SecurityContext" $ \v ->
        SecurityContext
            <$> v A..:? "runAsNonRoot"
            <*> v A..:? "runAsUser"

-- | Resource requirements
data Resources = Resources
    { requests :: Maybe ResourceRequirements
    , limits   :: Maybe ResourceRequirements
    } deriving (Eq, Show, Generic)

instance FromJSON Resources where
    parseJSON = A.withObject "Resources" $ \v ->
        Resources
            <$> v A..:? "requests"
            <*> v A..:? "limits"

-- | Resource requirements (CPU, memory)
data ResourceRequirements = ResourceRequirements
    { cpu    :: Maybe Text
    , memory :: Maybe Text
    } deriving (Eq, Show, Generic)

instance FromJSON ResourceRequirements where
    parseJSON = A.withObject "ResourceRequirements" $ \v ->
        ResourceRequirements
            <$> v A..:? "cpu"
            <*> v A..:? "memory"
